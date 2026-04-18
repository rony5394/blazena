package host

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	cfg "github.com/rony5394/blazena/config"
);

func exchangeKeys(Config cfg.Config, sshKeyPem string)string{
	var body struct{
		SshPkPem string `json:"sshPkPem"` 
	} = struct{SshPkPem string `json:"sshPkPem"`}{
		SshPkPem: sshKeyPem,
	};

	bodyEncoded, err := json.Marshal(body);

	if err != nil {
		panic("Failed to marshal body."+ err.Error());
		slog.Error("Failed to marshal body.", slog.Any("propagatedError", err), slog.String("note", "Input for this marshal operation is that ssh pk. So the kebab is going on!"))
		os.Exit(42);
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/keys", bytes.NewBuffer(bodyEncoded)); 

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

	rsBodyRaw, err := io.ReadAll(rs.Body);

	if err != nil{
		slog.Error("Failed to read response body!" , slog.Any("propagatedError", err));
		os.Exit(1);
	}

	var rsBody struct{
		HostPkPem string `json:"hostPkPem"` 
	};

	err = json.Unmarshal(rsBodyRaw, &rsBody);
	if err != nil{
		slog.Error("Failed to unmarshal rsBodyRaw!", slog.Any("propagatedError", err));
		os.Exit(1);
	}


	return rsBody.HostPkPem;
}


