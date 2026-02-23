package docker

import (
	"context"
	"io"
	"net/http"
	"time"

	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/swarm"
);

func scaleDown(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}


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

	serviceId := bodyDecoded.ServiceId;

	inspectresoult, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), serviceId, swarm.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	if inspectresoult.Spec.Labels["blazena.scaledDown"] != "" {
		fmt.Println("Tried to scale down already scaled down service! "+ serviceId);
		return;
	}

	originalScale := inspectresoult.Spec.Mode.Replicated.Replicas;
	updatedSpec := inspectresoult.Spec;

	newScale := uint64(0);
	updatedSpec.Mode.Replicated.Replicas = &newScale;
	updatedSpec.Labels["blazena.scaledDown"] = "true";

	scale.Store(serviceId, *originalScale);

	_, err = ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{}); 
	if(err != nil){
		panic("Failed to update service."+ err.Error());
	}

	//TODO: Add proper wait system
	time.Sleep(15 * time.Second);
}

func scaleUp(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}


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

	serviceId := bodyDecoded.ServiceId;
	inspectresoult, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), serviceId, swarm.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	if inspectresoult.Spec.Labels["blazena.scaledDown"] != "true" {
		fmt.Println("Tried to scale up service that was not scaled down by blazena! "+serviceId);
		return;
	}

	originalScale, ok := scale.Load(serviceId);  
	if(!ok){
		panic("Its not okay!");
	}

	originalScaleChecked, ok := originalScale.(uint64);
	if(!ok){
		panic("Its very not okay!")
	}
	updatedSpec := inspectresoult.Spec;

	updatedSpec.Mode.Replicated.Replicas = &originalScaleChecked;
	delete(updatedSpec.Labels, "blazena.scaledDown");

	ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{});

	//TODO: Add proper wait system
	time.Sleep(15 * time.Second);
}
