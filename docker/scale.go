package docker

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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
	updatedSpec.Labels["blazena.originalScale"] = strconv.FormatUint(*originalScale, 10);

	_, err = ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{}); 
	if(err != nil){
		panic("Failed to update service."+ err.Error());
	}
	ctx, cancel := context.WithTimeout(context.Background(), theConfig.Constants.ServiceScaleTimeout);
	defer cancel();	

	waitForScale(serviceId, ctx, 0);
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

	originalScale := inspectresoult.Spec.Labels["blazena.originalScale"];

	if(originalScale == ""){
		panic("Its not okay!");
	}

	originalScaleChecked, err := strconv.ParseUint(originalScale, 10, 64);
	if(err != nil){
		panic("Its very not okay!"+ err.Error())
	}
	updatedSpec := inspectresoult.Spec;

	updatedSpec.Mode.Replicated.Replicas = &originalScaleChecked;
	delete(updatedSpec.Labels, "blazena.scaledDown");
	delete(updatedSpec.Labels, "blazena.originalScale");

	_, err = ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{});

	if err != nil {
		slog.Error("Failed to update/scale a service.", slog.Any("propagatedError", err), slog.String("serviceId", serviceId));
		os.Exit(1);
	}

	ctx, cancel := context.WithTimeout(context.Background(), theConfig.Constants.ServiceScaleTimeout);
	defer cancel();	

	waitForScale(serviceId, ctx, int(originalScaleChecked));
}

func waitForScale(serviceId string, ctx context.Context, desiredCount int){
	startTime := time.Now();
	for ctx.Err() == nil {
		tasks, err := ApiClient.TaskList(context.Background(), swarm.TaskListOptions{});

		if err != nil {
			slog.Error("Failed to list tasks.", slog.Any("propagatedError", err));
			os.Exit(1);
		}

		var running int;
		for _, task := range tasks {
			if task.ServiceID != serviceId {
				continue;
			}

			if task.Status.State == swarm.TaskStateRunning{
				running ++;
			}
		}

		if running == desiredCount {
			slog.Debug("Rescaled Service",
				slog.String("serviceId", serviceId),
				slog.Any("took", time.Since(startTime)),
				slog.Any("targetScale", desiredCount),
			);
			return;
		}
		time.Sleep(1*time.Second);
	}
	if ctx.Err() == context.DeadlineExceeded{
		slog.Error("Failed to rescale service in given time.", slog.Any("serviceId", serviceId));
	}
}
