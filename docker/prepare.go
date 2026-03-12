package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encoding/base64"
	"time"
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/image"
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

	io.Copy(os.Stdout, ipc);


	labels := inspectResoults.Spec.Labels;

	maxConcurrent := uint64(1);
	totalCompletions := uint64(1);
	stopGracePeriod := time.Second * 5;
	targetNode := labels["blazena.node"];
	helperCommand := `ssh-keygen -t ed25519 -f /host_key && \
			echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIByYbl8vu946LPycSO5pBohq3vMvvl+wX7snu1Bqpd7p test" > /root/.ssh/authorized_keys && \
			/usr/sbin/sshd -h /host_key -p 2222 -D`;

	
	_, err = ApiClient.ServiceCreate(context.Background(), swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: "BlazenaHelper",
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
						Source: bodyDecoded.VolumeId,
						Target: "/volume",
						Type: "volume",
						ReadOnly: true,
					},
				},
				StopGracePeriod: &stopGracePeriod,
			},
			Placement: &swarm.Placement{
				Constraints: []string{"node.hostname=="+targetNode},
			},
			Networks: []swarm.NetworkAttachmentConfig{swarm.NetworkAttachmentConfig{
				Target: "blazenaPohar",
			}},
		},
	}, swarm.ServiceCreateOptions{});

	if err != nil {
		panic("Failed to create helper service."+ err.Error());
	}

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

