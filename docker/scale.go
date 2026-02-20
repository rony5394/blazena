package docker

import (
	"context"

	"github.com/docker/docker/api/types/swarm"
);

func scaleDown(serviceId string){
	inspectresoult, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), serviceId, swarm.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	originalScale := inspectresoult.Spec.Mode.Replicated.Replicas;
	updatedSpec := inspectresoult.Spec;

	newScale := uint64(0);
	updatedSpec.Mode.Replicated.Replicas = &newScale;

	scale.Store(serviceId, *originalScale);

	_, err = ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{}); 
	if(err != nil){
		panic("Failed to update service."+ err.Error());
	}
}

func scaleUp(serviceId string){
	inspectresoult, _, err := ApiClient.ServiceInspectWithRaw(context.Background(), serviceId, swarm.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
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

	ApiClient.ServiceUpdate(context.Background(), serviceId, inspectresoult.Version, updatedSpec, swarm.ServiceUpdateOptions{});
}
