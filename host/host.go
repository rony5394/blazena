package host

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	cfg "github.com/rony5394/blazena/config"
	"github.com/rony5394/blazena/shared"
)

var token string = "12345";

type aService struct{
	ServiceId string `json:"serviceId"`;
	VolumeNames []string `json:"volumeNames"`;
	Node string `json:"node"`;
}

func Run(Config cfg.Config) {
	DockerClient, err := client.NewClientWithOpts(client.FromEnv);
	if err != nil {
		panic("Failed to create DockerClient.");
	}

	_, err = DockerClient.Ping(context.Background())
	if err != nil {
		panic("Failed to ping DockerClient.");
	}

	sshKeyPair := shared.GenerateSSHKeypair();
	sshHostPkPem := exchangeKeys(Config, string(sshKeyPair.Public));
	createStorageContainer(Config, DockerClient, sshKeyPair.Private, sshHostPkPem);

	services := getServices(Config);

	for _, service := range services {
		fmt.Println("Scaling Down: "+service.ServiceId)
		scale(Config, service.ServiceId, false);
		fmt.Println("Done!");
		for _, volume := range service.VolumeNames{
			fmt.Println("Preparing: " + service.ServiceId + " volume: " + volume);
			if !prepareService(Config, service, volume) {continue}
			fmt.Println("Done!");

			storagePath, _ := generateStoragePath(Config, service.Node, volume, DockerClient);
			fmt.Println(storagePath);
			command := `rsync -avz --delete -e "ssh -i /ssh-key -p 2222 -o StrictHostKeyChecking=yes -o UserKnownHostsFile=/expected-host-key" \
				root@tasks.`+ Config.Constants.HelperServiceName +`:/volume/ ` +storagePath;

			exec, err := DockerClient.ContainerExecCreate(context.Background(), Config.Constants.StorageContainerName, container.ExecOptions{
				Cmd: []string{"sh", "-c", command},
				AttachStdout: true,
				AttachStderr: true,
				Tty: false,
			});
			if err != nil {
				panic("Failed to create rsync exec!"+err.Error());
			}


			resp, err := DockerClient.ContainerExecAttach(context.Background(), exec.ID, container.ExecStartOptions{});
			defer resp.Close();

			io.Copy(os.Stdout, resp.Reader)

			time.Sleep(30*time.Second);
			fmt.Println("Cleaning Up: " + service.ServiceId);
			cleanupService(Config, service);
			fmt.Println("Done!");
		}
		fmt.Println("Scaling up: "+service.ServiceId);
		scale(Config, service.ServiceId, true);
		fmt.Println("Done!");
	}

	DockerClient.ContainerRemove(context.Background(), Config.Constants.StorageContainerName, container.RemoveOptions{
		Force: true,
	});

	if !shutdown(Config){panic("Failed to shutdown docker api!");}
}

func getServices(Config cfg.Config)[]aService{
	req, err := http.NewRequest("GET", Config.DockerManagerBaseUrl + "/services", nil); 
	if err != nil {
		panic("Failed to create request."+ err.Error());
	}

	req.Header.Add("Authorization", "Bearer "+ token);

	res, err := http.DefaultClient.Do(req);
	if err != nil {
		panic("Failed to send request."+ err.Error());
	}
	reader, err := io.ReadAll(res.Body);
	if err != nil {
		panic("Failed to decode response body."+err.Error());
	}
	var services []aService; 
	err = json.Unmarshal(reader, &services);
	if err != nil {
		panic("Failed to unmarshal response.");
	}
	return services;
}

func cleanupService(Config cfg.Config, service aService)bool{
	_, ok := Config.Nodes[service.Node];
	if !ok {
		fmt.Println("Node", service.Node, "refferenced in", service.ServiceId ,"service does not exists!");
		return false;
	}
	
	var body struct{
		ServiceId string `json:"serviceId"`
		VolumeId string `json:"volumeId"`
	} = struct{ServiceId string "json:\"serviceId\""; VolumeId string "json:\"volumeId\""}{
			ServiceId: service.ServiceId,
		}

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/cleanup", bytes.NewBuffer(bodyEncoded)); 

	if err != nil{
		panic("Failed to create http request"+ err.Error());
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	rs, err := http.DefaultClient.Do(rq);
	defer rs.Body.Close();

	if err != nil{
		panic("Failed to send http request"+ err.Error());
	}

	return true;

}

func shutdown(Config cfg.Config)bool{
	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/shutdown", nil); 

	if err != nil{
		panic("Failed to create http request"+ err.Error());
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	_, err = http.DefaultClient.Do(rq);

	// if err != nil{
	// 	panic("Failed to send http request"+ err.Error());
	// }

	return true;

}

func addToTar(tw *tar.Writer, filename string, content string) error{
	hdr := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len([]byte(content))),
	};

	if err := tw.WriteHeader(hdr); err != nil{
		return err;
	}

	_, err := tw.Write([]byte(content))

	return err;
}

func createIfMissing(targetPath string, DockerClient *client.Client, cfg cfg.Config) error{
	const cmd = `#!/bin/sh
	set -e

	TARGET_PATH=$1

	# Remove trailing slash
	TARGET_PATH=${TARGET_PATH%/}

	CURRENT=""

	case "$TARGET_PATH" in
	    /*) CURRENT="/" ;;
	esac

	OLD_IFS=$IFS
	IFS='/'

	for PART in $TARGET_PATH; do
	    [ -z "$PART" ] && continue

	    if [ "$CURRENT" = "/" ]; then
		NEXT="${CURRENT}${PART}"
	    else
		NEXT="${CURRENT}/${PART}"
	    fi

	    if [ ! -e "$NEXT" ]; then
		case "$PART" in
		    @*)
			echo "Creating Btrfs subvolume: $NEXT"
			btrfs subvolume create "$NEXT"
			;;
		    *)
			echo "Creating directory: $NEXT"
			mkdir "$NEXT"
			;;
		esac
	    else
		echo "Already exists: $NEXT"
	    fi

	    CURRENT="$NEXT"
	done

	IFS=$OLD_IFS`;

	exec, err := DockerClient.ContainerExecCreate(context.Background(), cfg.Constants.StorageContainerName, container.ExecOptions{
		Cmd: []string{"sh", "-c", cmd, "_", targetPath},
		AttachStdout: true,
		AttachStderr: true,
		Tty: false,
	});

	if err != nil {
		panic("Failed to create execute!"+err.Error());
	}

	resp, err := DockerClient.ContainerExecAttach(context.Background(), exec.ID, container.ExecStartOptions{});
	defer resp.Close();

	if err != nil {
		panic("Failed to atach to exec!"+err.Error());
	}

	inspect, err := DockerClient.ContainerExecInspect(context.Background(), exec.ID);

	if(inspect.ExitCode != 0){
		fmt.Println("<resp>");
		io.Copy(os.Stdout, resp.Reader);
		fmt.Println("</resp>");
		return errors.New("Execution did return non zero code!"); 
	}

	return nil;
}

func generateStoragePath(cfg cfg.Config, node string, volumeId string, DockerClient *client.Client) (string, error){
	var path string;
	path += "/volume";

	path += "/@"+ node +"/@"+ volumeId;

	err := createIfMissing(path, DockerClient, cfg);

	if err != nil {
		return "", err;
	}

	return path, nil;
}
