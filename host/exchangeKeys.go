package host

import (
	"encoding/json"
	"io"
	"net/http"
	"bytes"
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
	}

	rq, err := http.NewRequest("POST", Config.DockerManagerBaseUrl + "/keys", bytes.NewBuffer(bodyEncoded)); 

	if err != nil{
		panic("Failed to create http request"+ err.Error());
	}

	rq.Header.Set("Authorization", "Bearer "+ token);
	rq.Close = true;
	rs, err := http.DefaultClient.Do(rq);

	if err != nil{
		panic("Failed to send http request"+ err.Error());
	}

	defer rs.Body.Close();

	rsBodyRaw, err := io.ReadAll(rs.Body);

	if err != nil{
		panic("Failed to read response's body!"+err.Error());
	}

	var rsBody struct{
		HostPkPem string `json:"hostPkPem"` 
	};

	err = json.Unmarshal(rsBodyRaw, &rsBody);
	if err != nil{
		panic("Failed to unmarshal rsBodyRaw!"+ err.Error());
	}


	return rsBody.HostPkPem;
}


