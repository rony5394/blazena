package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/moby/moby/client"
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

	scaleUp(bodyDecoded.ServiceId);

	listResoult, err := ApiClient.ServiceList(context.Background(), client.ServiceListOptions{});
	if err != nil {
		panic("Failed to list services."+ err.Error());
	}

	var helperServiceId string;

	for _, service := range listResoult.Items {
		if service.Spec.Labels["blazena.helper"] != "true" {
			continue;
		}
		helperServiceId = service.ID;
		break;
	}

	if helperServiceId == ""{
		panic("Helper service not found!");
	}

	_, err = ApiClient.ServiceRemove(context.Background(), helperServiceId, client.ServiceRemoveOptions{});
	if err != nil {
		panic("Failed to remove helper service."+ err.Error());
	}
	fmt.Fprint(w, bodyDecoded.ServiceId);
}
