package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/docker/docker/api/types/swarm"

	"github.com/rony5394/blazena/shared"
)

func exchangeKeys(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprint(w, "Method Not Allowed");
		return;
	}
	if !bearerAuth(w, r) {return;}

	rawBody, err := io.ReadAll(r.Body);
	if err != nil {
		panic("Failed to read body!");
	}

	var bodyDecoded struct{
		SshPkPem string `json:"sshPkPem"`
	};

	err = json.Unmarshal(rawBody, &bodyDecoded);
	if err != nil {
		panic("Failed to unmarshal json."+ err.Error());
	}
	sshClientPkPem := bodyDecoded.SshPkPem;
	hostKeypair := shared.GenerateSSHKeypair();

	encoded, err := json.Marshal(struct{HostPkPem string `json:"hostPkPem"`}{HostPkPem: hostKeypair.Public});
	if err != nil {
		slog.Error("Failed to marshal host pk into response.", slog.Any("propagatedError", err));
		os.Exit(42);
	}

	_, err = ApiClient.ConfigCreate(context.Background(), swarm.ConfigSpec{
		Data: []byte(sshClientPkPem), 
		Annotations: swarm.Annotations{Name: theConfig.Constants.SSHClientPKConfigName},
	});

	if err != nil {
		slog.Error("Failed to create a config.", slog.Any("propagatedError", err));
		os.Exit(1);
	}

	_, err = ApiClient.SecretCreate(context.Background(), swarm.SecretSpec{
		Data: []byte(hostKeypair.Private),
		Annotations: swarm.Annotations{Name: theConfig.Constants.SSHHostSKSecretName}, 
	});

	if err != nil {
		slog.Error("Failed to create a secret.", slog.Any("propagatedError", err));
		os.Exit(1);
	}


	fmt.Fprint(w, string(encoded));
}
