package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
)

func prepare(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}

	//TODO: add token auth

	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!");
	}

	var bodyDecoded struct{
		volume string
		node string
	};

	err = json.Unmarshal(rawBody, &bodyDecoded);
	if err != nil {
		panic("Failed to unmarshal json."+ err.Error());
	}

	maxConcurrent := uint64(1);
	totalCompletions := uint64(1);
	targetVolume := bodyDecoded.volume;
	targetNode := bodyDecoded.node;
	
	ApiClient.ServiceCreate(context.Background(), client.ServiceCreateOptions{
		Spec: swarm.ServiceSpec{
			Mode: swarm.ServiceMode{
				ReplicatedJob: &swarm.ReplicatedJob{
					MaxConcurrent: &maxConcurrent,
					TotalCompletions: &totalCompletions,
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image: "docker.io/library/alpine:latest",
					Mounts: []mount.Mount{
						mount.Mount{
							Source: targetVolume,
							Target: "/volume",
							Type: "bind",
						},
					},
				},
				Placement: &swarm.Placement{
					Constraints: []string{"node.hostname=="+targetNode},
				},
			},
		},
	});
}
