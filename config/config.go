package config;

type Config struct {
	Nodes map[string] struct{
		Ip string
		DockerVolumePath string
	}
	DockerManagerBaseUrl string
	LocalBasePath string
};
