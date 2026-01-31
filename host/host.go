package host

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cfg "github.com/rony5394/blazena/config"
)

var token string = "12345";

type aService struct{
	ServiceId string `json:"serviceId"`;
	VolumeNames []string `json:"volumeNames"`;
	Node string `json:"node"`;
}

func Run(Config cfg.Config) {
	services := getServices(Config);

	for _, service := range services {
		_, ok := Config.Nodes[service.Node];
		if !ok {
			fmt.Println("Node", service.Node, "refferenced in", service.ServiceId ,"service does not exists!");
			continue;
		}
		for _, volume := range service.VolumeNames{
			remoteFilePath := Config.Nodes[service.Node].DockerVolumePath + "/" + volume;
			remotePath := Config.User + "@" + Config.Nodes[service.Node].Ip + ":" + remoteFilePath;
			base := "rsync -avh --delete --progress -e 'ssh -i ./ssh-key' ";
			localPath := Config.LocalBasePath + "/@current/@" + service.Node;

			fmt.Println(base + remotePath + " " + localPath);

		}
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
