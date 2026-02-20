package host

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/client"
	cfg "github.com/rony5394/blazena/config"
)

var token string = "12345";
var DockerClient *client.Client;

type aService struct{
	ServiceId string `json:"serviceId"`;
	VolumeNames []string `json:"volumeNames"`;
	Node string `json:"node"`;
}

func Run(Config cfg.Config) {
	DockerClient, err := client.NewClientWithOpts(client.FromEnv);
	if err != nil {
		panic("Failed to create DockerClient.");
	}

	_, err = DockerClient.Ping(context.Background())
	if err != nil {
		panic("Failed to ping DockerClient.");
	}

	services := getServices(Config);

	for _, service := range services {
		_, ok := Config.Nodes[service.Node];
		if !ok {
			fmt.Println("Node", service.Node, "refferenced in", service.ServiceId ,"service does not exists!");
			continue;
		}
		
		var body struct{
			ServiceId string `json:"serviceId"`
		} = struct{ServiceId string "json:\"serviceId\""}{
				ServiceId: service.ServiceId,
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

		os.Exit(0);

	}
}

func getServices(Config cfg.Config)[]aService{
	req, err := http.NewRequest("GET", Config.DockerManagerBaseUrl + "/services", nil); 
	if err != nil {
		panic("Failed to create request."+ err.Error());
	}

	req.Header.Add("Authorization", "Bearer "+ token);

	res, err := http.DefaultClient.Do(req);
	if err != nil {
		panic("Failed to send request."+ err.Error());
	}
	reader, err := io.ReadAll(res.Body);
	if err != nil {
		panic("Failed to decode response body."+err.Error());
	}
	var services []aService; 
	err = json.Unmarshal(reader, &services);
	if err != nil {
		panic("Failed to unmarshal response.");
	}
	return services;
}
