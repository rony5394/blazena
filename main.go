package main

import (
	"os"
	"github.com/rony5394/blazena/docker"
	"github.com/rony5394/blazena/host"
	cfg "github.com/rony5394/blazena/config"
);

func main() {
	if(len(os.Args) < 2){
		panic("Usage: blazena <mode>");
	}

	var config = cfg.GetConfig();

	mode := os.Args[1];
	switch mode {
		case "docker":
			docker.Run(config);
		break;
		case "host":
			host.Run(config);	
		break;
		default:
			panic("Invalid runtime mode!");
	}
}
