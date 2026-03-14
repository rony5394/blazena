package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"context"
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
	sshPkPem := bodyDecoded.SshPkPem;
	hostKeypair := shared.GenerateSSHKeypair();
	encoded, err := json.Marshal(struct{HostPkPem string `json:"hostPkPem"`}{HostPkPem: hostKeypair.Public});
	if err != nil {
		panic("I wonder how. I wonder why?"+err.Error());
	}

	ApiClient.ConfigCreate(context.Background(), swarm.ConfigSpec{
		Data: []byte(sshPkPem), 
		Annotations: swarm.Annotations{Name: "blazenaSSHPublicKey"},
	});

	ApiClient.SecretCreate(context.Background(), swarm.SecretSpec{
		Data: []byte(hostKeypair.Private),
		Annotations: swarm.Annotations{Name: "blazenaSSHHostPrivateKey"},
	});


	fmt.Fprint(w, string(encoded));
}
