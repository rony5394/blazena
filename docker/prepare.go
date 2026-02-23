package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
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

	scaleDown(bodyDecoded.ServiceId);
	//TODO: Add proper wait system
	time.Sleep(10*time.Second);

	inspectResoults, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), bodyDecoded.ServiceId, swarm.ServiceInspectOptions{});
	if err != nil{
		panic("Failed to inspect service."+ err.Error());
	}

	labels := inspectResoults.Spec.Labels;
	time.Sleep(10);

	maxConcurrent := uint64(1);
	totalCompletions := uint64(1);
	targetNode := labels["blazena.node"];
	helperCommand := `ssh-keygen -t ed25519 -f /host_key && \
			echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIByYbl8vu946LPycSO5pBohq3vMvvl+wX7snu1Bqpd7p test" > /root/.ssh/authorized_keys && \
			/usr/sbin/sshd -h /host_key -p 22 -D`;

	
	_, err = ApiClient.ServiceCreate(context.Background(), swarm.ServiceSpec{
		Annotations: swarm.Annotations{
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
				Image: "registry.lan.ronycloud.dev/blazena/helper:latest",
				Command: []string{"sh", "-c", helperCommand},
				Mounts: []mount.Mount{
					mount.Mount{
						Source: bodyDecoded.VolumeId,
						Target: "/volume",
						Type: "volume",
					},
				},
			},
			Placement: &swarm.Placement{
				Constraints: []string{"node.hostname=="+targetNode},
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				swarm.PortConfig{
					Protocol: swarm.PortConfigProtocolTCP,
					TargetPort: uint32(22),	
					PublishedPort: uint32(2222),
					PublishMode: swarm.PortConfigPublishModeHost,
				},
			},
		},
	}, swarm.ServiceCreateOptions{});

	if err != nil {
		panic("Failed to create helper service."+ err.Error());
	}

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

