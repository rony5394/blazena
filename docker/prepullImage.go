package docker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/swarm"
)

func prepullImage(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}

	if !bearerAuth(w, r){return;}

	sc, err := ApiClient.ServiceCreate(context.Background(), swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: theConfig.Constants.PrepullImageServiceName, 
			Labels: map[string]string{"blazena.prepull": "true"},
		},
		Mode: swarm.ServiceMode{
			GlobalJob: &swarm.GlobalJob{},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: theConfig.BlazenaImageUrl, 
				Command: []string{"sleep", "1s"},
			},
		},
	}, swarm.ServiceCreateOptions{
		QueryRegistry: true,
		EncodedRegistryAuth: theConfig.EncodedRegistryAuth,
	});

	if err != nil {
		slog.Error("Failed to create prepull service", slog.Any("propagatedError", err));
		os.Exit(3);
	}

	slog.Info("Started Prepull of blazena image.");

	go func(){
	CHECKLOOP: for {
		time.Sleep(3*time.Second);
		tasks, err := ApiClient.TaskList(context.Background(), swarm.TaskListOptions{});

		if err != nil {
			slog.Error("Failed to list tasks.", slog.Any("propagatedError", err));
			os.Exit(3);
		}

		for _, task := range tasks {
			if task.ServiceID != sc.ID {continue};

			if task.Status.State != "complete" {
				time.Sleep(1*time.Second);
				continue CHECKLOOP
			};
		}

		err = ApiClient.ServiceRemove(context.Background(), sc.ID);

		if err != nil {
			slog.Warn("Failed to remove prepull service.", slog.Any("propagatedError", err));
		}

		slog.Info("Prepull Finished");
		break CHECKLOOP;
	}
	}();
}
