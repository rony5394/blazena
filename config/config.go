package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/docker/docker/api/types/registry"
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
	EncodedRegistryAuth string
	Constants struct{
		OverlayNetworkName string
		HelperServiceName string
		StorageContainerName string
		PrepullImageServiceName string
		ServiceScaleTimeout time.Duration
		SSHClientPKConfigName string
		SSHHostSKSecretName string
	}
}

type RegistryAuth struct {
	Username string
	Password string
}

func GetConfig()(Config, error){
	var cfg Config;

	rawConfig, err := os.ReadFile("/config.json");
	if err != nil{
		return cfg, errors.New("Failed it load config file." + err.Error());
	}

	// Set defaults
	cfg.Constants.OverlayNetworkName = "blazenaPohar";
	cfg.Constants.HelperServiceName = "blazenaHelper";
	cfg.Constants.StorageContainerName = "blazenaStorage";
	cfg.Constants.PrepullImageServiceName = "blazenaPrepull";
	cfg.Constants.ServiceScaleTimeout = time.Second * 15; 
	cfg.Constants.SSHClientPKConfigName = "blazenaSSHClientPublicKey";
	cfg.Constants.SSHHostSKSecretName = "blazenaSSHHostPrivateKey";

	err = json.Unmarshal(rawConfig, &cfg);

	if err != nil{
		return cfg, errors.New("Failed to unmarshal config: " + err.Error());
	}

	authConfig := registry.AuthConfig{
		Username: cfg.RegistryAuth.Username, 
		Password: cfg.RegistryAuth.Password,
	}

	authJSON, err := json.Marshal(authConfig)

	if err != nil {
		panic("Failed to marshal auth config!"+ err.Error());
	}

	cfg.EncodedRegistryAuth = base64.StdEncoding.EncodeToString(authJSON);


	return cfg, err;
}
