package host

import(
	"fmt"
	"encoding/json"
	"bytes"
	cfg "github.com/rony5394/blazena/config"
	"net/http"
);

func prepareService(Config cfg.Config, service aService, targetVolume string) bool{
	_, ok := Config.Nodes[service.Node];
	if !ok {
		fmt.Println("Node", service.Node, "refferenced in", service.ServiceId ,"service does not exists!");
		return false;
	}
	
	var body struct{
		ServiceId string `json:"serviceId"`
		VolumeId string `json:"volumeId"`
	} = struct{ServiceId string "json:\"serviceId\""; VolumeId string "json:\"volumeId\""}{
			ServiceId: service.ServiceId,
			VolumeId: targetVolume,
		}

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/prepare", bytes.NewBuffer(bodyEncoded)); 

	if err != nil{
		panic("Failed to create http request"+ err.Error());
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	rs, err := http.DefaultClient.Do(rq);
	defer rs.Body.Close();

	if err != nil{
		panic("Failed to send http request"+ err.Error());
	}

	return true;
}


