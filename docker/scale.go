package docker

import (
	"context"

	"github.com/moby/moby/client"
);

func scaleDown(serviceId string){
	inspectresoult, err := ApiClient.ServiceInspect(context.Background(), serviceId, client.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	originalScale := inspectresoult.Service.Spec.Mode.Replicated.Replicas;
	updatedSpec := inspectresoult.Service.Spec;

	newScale := uint64(0);
	updatedSpec.Mode.Replicated.Replicas = &newScale;

	scale.Store(serviceId, *originalScale);

	_, err = ApiClient.ServiceUpdate(context.Background(), serviceId, client.ServiceUpdateOptions{
		Spec: updatedSpec,
		Version: inspectresoult.Service.Version,
	});

	if(err != nil){
		panic("Failed to update service."+ err.Error());
	}
}

func scaleUp(serviceId string){
	inspectresoult, err := ApiClient.ServiceInspect(context.Background(), serviceId, client.ServiceInspectOptions{});

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
	updatedSpec := inspectresoult.Service.Spec;

	updatedSpec.Mode.Replicated.Replicas = &originalScaleChecked;


	ApiClient.ServiceUpdate(context.Background(), serviceId, client.ServiceUpdateOptions{
		Spec: updatedSpec,
		Version: inspectresoult.Service.Version,
		
	});

}
