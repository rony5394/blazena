package docker 

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type Keypair struct {
	public []byte
	private []byte
}

func generateKeypair() Keypair {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	sshPubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		panic(err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPubKey)

	return Keypair{
		private: privPem,
		public: pubBytes,
	};

}
