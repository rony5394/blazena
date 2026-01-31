package main

import (
	"encoding/json"
	"os"

	"github.com/rony5394/blazena/docker"
	"github.com/rony5394/blazena/host"
	cfg "github.com/rony5394/blazena/config"
);

var Config cfg.Config = cfg.Config{
	Nodes: make(map[string]struct{Ip string; DockerVolumePath string}),
};


func main() {
	if(len(os.Args) < 2){
		panic("Usage: blazena <mode>");
	}

	rawConfig, err := os.ReadFile("./config.json");
	if err != nil{
		panic("Failed it load config file." + err.Error());
	}

	err = json.Unmarshal(rawConfig, &Config);

	if err != nil{
		panic("Failed to unmarshal config." + err.Error())
	}

	mode := os.Args[1];
	switch mode {
		case "docker":
			docker.Run();
		break;
		case "host":
			host.Run(Config);	
		break;
		default:
			panic("Invalid runtime mode!");
	}
}
