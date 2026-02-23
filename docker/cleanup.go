package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!");
	}

	var bodyDecoded struct{
		ServiceId string `json:"serviceId"`
	};

	err = json.Unmarshal(rawBody, &bodyDecoded);
	if err != nil {
		panic("Failed to unmarshal json."+ err.Error());
	}

	listResoult, err := ApiClient.ServiceList(context.Background(), swarm.ServiceListOptions{});
	if err != nil {
		panic("Failed to list services."+ err.Error());
	}

	var helperServiceId string;

	for _, service := range listResoult{
		if service.Spec.Labels["blazena.helper"] != "true" {
			continue;
		}
		helperServiceId = service.ID;
		break;
	}

	if helperServiceId == ""{
		panic("Helper service not found!");
	}

	err = ApiClient.ServiceRemove(context.Background(), helperServiceId);
	if err != nil {
		panic("Failed to remove helper service."+ err.Error());
	}
	time.Sleep(15*time.Second);

	fmt.Fprint(w, bodyDecoded.ServiceId);
}
