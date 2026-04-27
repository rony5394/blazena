package host

import (
	"bytes"
	"log/slog"
	"net/http"
	"os"

	cfg "github.com/rony5394/blazena/config"
);

func prepullImage(Config cfg.Config){
	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/prepull", bytes.NewBufferString("{}")); 

	if err != nil{
		slog.Error("Failed to create request", slog.Any("propagatedError", err), slog.String("note", "not send just create the object"));
		os.Exit(1);
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	rs, err := http.DefaultClient.Do(rq);

	if err != nil{
		slog.Error("Failed to send http request", slog.Any("propagatedError", err));
		os.Exit(1);
	}

	defer rs.Body.Close();
}
