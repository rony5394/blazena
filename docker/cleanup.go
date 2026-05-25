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

func cleanup(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}

	if !bearerAuth(w, r) {return;}

	listResoult, err := ApiClient.ServiceList(context.Background(), swarm.ServiceListOptions{});
	if err != nil {
		slog.Error("Failed to list services", slog.Any("propagatedError", err));
		os.Exit(1);
	}

	var helperServiceId string;
	var helperServices int;

	for _, service := range listResoult{
		if service.Spec.Labels["blazena.helper"] != "true" {
			continue;
		}
		helperServiceId = service.ID;
		helperServices ++;
	}

	if helperServiceId == ""{
		slog.Warn("Helper service wasn't found");
		http.Error(w, "Internal Server Error", http.StatusInternalServerError);
		return;
	}

	if helperServices > 1{
		slog.Error("There are more than 1 helper service.");
		os.Exit(1);
	}

	err = ApiClient.ServiceRemove(context.Background(), helperServiceId);
	if err != nil {
		panic("Failed to remove helper service."+ err.Error());
	}
	//TODO: add proper wait system
	time.Sleep(7*time.Second);

	fmt.Fprint(w, helperServiceId);
}
