package main

import (
	"os"
	"github.com/rony5394/blazena/docker"
	"github.com/rony5394/blazena/host"
	cfg "github.com/rony5394/blazena/config"
	"log/slog"
);

func main() {
	if(len(os.Args) < 2){
		panic("Usage: blazena <mode>");
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}));
	slog.SetDefault(logger);

	config, err := cfg.GetConfig();

	if(err != nil){
		slog.Error("Failed to load config!", slog.Any("propagatedError", err.Error()));
		os.Exit(1);
	}

	slog.Debug("Config", slog.Any("Value", config));

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
