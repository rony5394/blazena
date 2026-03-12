package config;

type Config struct {
	Nodes map[string] struct{
		Ip string
		DockerVolumePath string
	}
	DockerManagerBaseUrl string
	LocalBasePath string
	BlazenaImageUrl string
	RegistryAuth RegistryAuth
};

type RegistryAuth struct {
	Username string
	Password string
}
