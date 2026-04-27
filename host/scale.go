package host

import (
	cfg "github.com/rony5394/blazena/config"
	"encoding/json"
	"net/http"
	"bytes"
);

func scale(Config cfg.Config, serviceId string, up bool)bool{
	var body struct{
		ServiceId string `json:"serviceId"`
	} = struct{ServiceId string "json:\"serviceId\""}{ 
			ServiceId: serviceId,
		}

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
	}

	uri := "/scale";

	if up {
		uri += "/up";
	} else {
		uri += "/down";
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + uri, bytes.NewBuffer(bodyEncoded)); 

	if err != nil{
		panic("Failed to create http request"+ err.Error());
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	rs, err := http.DefaultClient.Do(rq);
	if err == nil {
	rs.Body.Close();
	}

	if err != nil{
		panic("Failed to send http request"+ err.Error());
	}

	return true;
}


