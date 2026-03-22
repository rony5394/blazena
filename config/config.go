package config;

import (
	"os"
	"encoding/json"
);

type Config struct {
	Nodes map[string] struct{
		Ip string
		DockerVolumePath string
	}
	DockerManagerBaseUrl string
	LocalBasePath string
	BlazenaImageUrl string
	RegistryAuth RegistryAuth
	Constants struct{
		OverlayNetworkName string
		HelperServiceName string
		StorageContainerName string
	}
}

type RegistryAuth struct {
	Username string
	Password string
}

func GetConfig() Config {
	var cfg Config;

	rawConfig, err := os.ReadFile("./config.json");
	if err != nil{
		panic("Failed it load config file." + err.Error());
	}

	cfg.Constants.OverlayNetworkName = "blazenaPohar";
	cfg.Constants.HelperServiceName = "blazenaHelper";
	cfg.Constants.StorageContainerName = "blazenaStorage";
	

	err = json.Unmarshal(rawConfig, &cfg);

	if err != nil{
		panic("Failed to unmarshal config." + err.Error())
	}

	return cfg;
}
