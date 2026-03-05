package host

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	cfg "github.com/rony5394/blazena/config"
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

	createStorageContainer(Config, DockerClient);

	services := getServices(Config);

	for _, service := range services {
		fmt.Println("Scaling Down: "+service.ServiceId)
		scale(Config, service.ServiceId, false);
		fmt.Println("Done!");
		for _, volume := range service.VolumeNames{
			fmt.Println("Preparing: " + service.ServiceId + " volume: " + volume);
			if !prepareService(Config, service, volume) {continue}
			fmt.Println("Done!");

			// Skiping Host Key Check is temporary.
			command := `apk add --no-cache rsync openssh-client && \
				rsync -avz --delete -e "ssh -i /ssh-key -p 2222 -o StrictHostKeyChecking=no" \
				root@tasks.BlazenaHelper:/volume/ /tmp/` + volume;


			exec, err := DockerClient.ContainerExecCreate(context.Background(), "BlazenaStorage", container.ExecOptions{
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

	DockerClient.ContainerRemove(context.Background(), "BlazenaStorage", container.RemoveOptions{
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

func prepareService(Config cfg.Config, service aService, targetVolume string) bool{
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
			VolumeId: targetVolume,
		}

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/prepare", bytes.NewBuffer(bodyEncoded)); 

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

func scale(Config cfg.Config, serviceId string, up bool)bool{
	var body struct{
		ServiceId string `json:"serviceId"`
	} = struct{ServiceId string "json:\"serviceId\""}{ 
			ServiceId: serviceId,
		}

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
	}

	uri := "/scale";

	if up {
		uri += "/up";
	} else {
		uri += "/down";
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + uri, bytes.NewBuffer(bodyEncoded)); 

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

func createStorageContainer(Config cfg.Config, DockerClient *client.Client){
	cr, err := DockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: "docker.io/library/alpine:3.3",
		Labels: map[string]string{
			"blazena.storage": "true",
		},
		Cmd: strslice.StrSlice{"sleep", "1h"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeBind,
					Source: Config.LocalBasePath,
					Target: "/volume",
					ReadOnly: true,
				},
			},
		//AutoRemove: true,
		NetworkMode: "blazenaPohar",
	}, &network.NetworkingConfig{
		}, &v1.Platform{}, "BlazenaStorage");

	if err != nil {
		panic("Failed to create BlazenaStorage container!"+err.Error());
	}

	err = DockerClient.ContainerStart(context.Background(), cr.ID, container.StartOptions{}); 
	
	if err != nil{
		panic("Failed to start BlazenaStorage container!"+err.Error());
	}


	reader, err := os.Open("./ssh-key.tar");

	if err != nil {
		panic("Failed To open ssh key file for reading!"+err.Error());
	}

	defer reader.Close();

	err = DockerClient.CopyToContainer(context.Background(), "BlazenaStorage", "/", reader, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	});
	if err != nil {
		panic("Failed to copy ssh key to container!"+err.Error());
	}


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
