package docker;
import(
	"net/http"
	"io"
	"encoding/json"
	"context"
	"github.com/moby/moby/client"
);

func scaleDown(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed);
		return;
	}

	if !bearerAuth(w, r) {return}

	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!" + err.Error());
	}

	var parsedBody struct{
		ServiceId string `json:"serviceId"`;
	};

	err = json.Unmarshal(rawBody, &parsedBody);
	if err != nil{
		panic("Failed to unmarshal request body!"+ err.Error());
	}

	inspectresoult, err := ApiClient.ServiceInspect(context.Background(), parsedBody.ServiceId, client.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	originalScale := inspectresoult.Service.Spec.Mode.Replicated.Replicas;
	updatedSpec := inspectresoult.Service.Spec;

	newScale := uint64(0);
	updatedSpec.Mode.Replicated.Replicas = &newScale;

	scale.Store(parsedBody.ServiceId, *originalScale);

	_, err = ApiClient.ServiceUpdate(context.Background(), parsedBody.ServiceId, client.ServiceUpdateOptions{
		Spec: updatedSpec,
		Version: inspectresoult.Service.Version,
	});

	if(err != nil){
		panic("Failed to update service."+ err.Error());
	}
}

func scaleUp(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed);
		return;
	}

	if !bearerAuth(w, r) {return}


	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!");
	}

	var parsedBody struct{
		ServiceId string `json:"serviceId"`;
	};

	err = json.Unmarshal(rawBody, &parsedBody);
	if err != nil{
		panic("Failed to unmarshal request body!"+ err.Error());
	}

	inspectresoult, err := ApiClient.ServiceInspect(context.Background(), parsedBody.ServiceId, client.ServiceInspectOptions{});

	if err != nil{
		panic("Error inspecting service!"+ err.Error());
	}

	originalScale, ok := scale.Load(parsedBody.ServiceId);  
	if(!ok){
		panic("Its not okay!");
	}

	originalScaleChecked, ok := originalScale.(uint64);
	if(!ok){
		panic("Its very not okay!")
	}
	updatedSpec := inspectresoult.Service.Spec;

	updatedSpec.Mode.Replicated.Replicas = &originalScaleChecked;


	ApiClient.ServiceUpdate(context.Background(), parsedBody.ServiceId, client.ServiceUpdateOptions{
		Spec: updatedSpec,
		Version: inspectresoult.Service.Version,
		
	});

}
