package shared 

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type Keypair struct {
	Public string 
	Private string 
}

func GenerateSSHKeypair() Keypair {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	privBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		panic(err)
	}

	privPem := pem.EncodeToMemory(privBlock)

	sshPubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		panic(err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPubKey)

	return Keypair{
		Private: string(privPem),
		Public:  string(pubBytes),
	}
}
