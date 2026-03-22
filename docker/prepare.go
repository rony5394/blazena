package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encoding/base64"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	cfg "github.com/rony5394/blazena/config"
)

func prepare(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}

	if !bearerAuth(w, r){return;}

	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!");
	}

	var bodyDecoded struct{
		ServiceId string `json:"serviceId"`
		VolumeId string `json:"volumeId"`
	};

	err = json.Unmarshal(rawBody, &bodyDecoded);
	if err != nil {
		panic("Failed to unmarshal json."+ err.Error());
	}

	inspectResoults, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), bodyDecoded.ServiceId, swarm.ServiceInspectOptions{});
	if err != nil{
		panic("Failed to inspect service."+ err.Error());
	}


	labels := inspectResoults.Spec.Labels;

	pullBlazenaImage();
	createHelper(theConfig, labels["blazena.node"], bodyDecoded.VolumeId);

	time.Sleep(7*time.Second);

	fmt.Fprint(w, bodyDecoded.ServiceId);
}

func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}

func getConfigIDByName(cli *client.Client, name string) (string, error) {
	ctx := context.Background()

	configs, err := cli.ConfigList(ctx, swarm.ConfigListOptions{}) 
	if err != nil {
		return "", err
	}

	for _, cfg := range configs {
		if cfg.Spec.Name == name {
			return cfg.ID, nil
		}
	}

	return "", fmt.Errorf("config not found: %s", name)
}

func getSecretIDByName(cli *client.Client, name string) (string, error) {
	ctx := context.Background()

	secrets, err := cli.SecretList(ctx, swarm.SecretListOptions{}) 
	if err != nil {
		return "", err
	}

	for _, sec := range secrets {
		if sec.Spec.Name == name {
			return sec.ID, nil
		}
	}

	return "", fmt.Errorf("config not found: %s", name)
}

func pullBlazenaImage(){
	authConfig := registry.AuthConfig{
		Username: theConfig.RegistryAuth.Username, 
		Password: theConfig.RegistryAuth.Password,
	}

	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic("Failed to marshal auth config!"+ err.Error());
	}

	authString := base64.URLEncoding.EncodeToString(authJSON);

	ipc, err := ApiClient.ImagePull(context.Background(), theConfig.BlazenaImageUrl, image.PullOptions{RegistryAuth: authString});
	if err != nil {
		panic("Failed to pull blazena image!"+ err.Error());
	}
	defer ipc.Close();

	io.Copy(io.Discard, ipc);
}

func createHelper(Config cfg.Config, targetNode string, targetVolume string){
	maxConcurrent := uint64(1);
	totalCompletions := uint64(1);
	stopGracePeriod := time.Second * 5;
	helperCommand := `/usr/sbin/sshd -h /host-key -p 2222 -D`;

	sshKeyConfigId, err := getConfigIDByName(ApiClient, "blazenaSSHPublicKey");

	if err != nil {
		panic("Docker needs both id and name to mount config for some reason and getting id of it failed!"+err.Error());
	}

	sshHostKeySecretId, err := getSecretIDByName(ApiClient, "blazenaSSHHostPrivateKey")
	_, err = ApiClient.ServiceCreate(context.Background(), swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: Config.Constants.HelperServiceName, 
			Labels: map[string]string{"blazena.helper": "true"},
		},
		Mode: swarm.ServiceMode{
			ReplicatedJob: &swarm.ReplicatedJob{
				MaxConcurrent: &maxConcurrent,
				TotalCompletions: &totalCompletions,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: theConfig.BlazenaImageUrl, 
				Command: []string{"sh", "-c", helperCommand},
				Mounts: []mount.Mount{
					mount.Mount{
						Source: targetVolume, 
						Target: "/volume",
						Type: "volume",
						ReadOnly: true,
					},
				},
				StopGracePeriod: &stopGracePeriod,
				Configs: []*swarm.ConfigReference{
					&swarm.ConfigReference{
						ConfigID: sshKeyConfigId, 
						ConfigName: "blazenaSSHPublicKey",
						File: &swarm.ConfigReferenceFileTarget{
							Name: "/root/.ssh/authorized_keys",
							Mode: 0600,
							UID: "0",
							GID: "0",
						},
					},
				},
				Secrets: []*swarm.SecretReference{
					&swarm.SecretReference{
						SecretID: sshHostKeySecretId,
						SecretName: "blazenaSSHHostPrivateKey",
						File: &swarm.SecretReferenceFileTarget{
							Name: "/host-key",
							Mode: 0600,
							UID: "0",
							GID: "0",
						},
					},
				},
			},
			Placement: &swarm.Placement{
				Constraints: []string{"node.hostname=="+targetNode},
			},
			Networks: []swarm.NetworkAttachmentConfig{swarm.NetworkAttachmentConfig{
				Target: Config.Constants.OverlayNetworkName,
			}},
		},
	}, swarm.ServiceCreateOptions{});

	if err != nil {
		panic("Failed to create helper service."+ err.Error());
	}

}
