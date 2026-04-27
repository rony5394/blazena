package main

import (
	"os"
	"github.com/rony5394/blazena/docker"
	"github.com/rony5394/blazena/host"
	cfg "github.com/rony5394/blazena/config"
	"log/slog"
);

/*
	If the exit code is X then it means Y:

	| X  | Y                                                                                        |
	|----|------------------------------------------------------------------------------------------|
	| 0  | Everything should be good, normal exit.                                                  |
	| 1  | Some common error, but that still mean it is going to crash.                             |
	| 3  | Ask yourself if you are not using dev version in prod. If not then spam the developer.   |
	| 42 | WHAT THE ACTUAL ***** IS HAPPENING. or assume something is very wrong in the app itself. |
	| 69 | [INSERT HERE]                                                                            |
*/

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
		case "pull":
			os.Exit(0);
		break;
		default:
			panic("Invalid runtime mode!");
	}
}
