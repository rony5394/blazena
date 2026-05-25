package shared

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"log/slog"
	"os"

	"golang.org/x/crypto/ssh"
)

type Keypair struct {
	Public string 
	Private string 
}

func GenerateSSHKeypair() Keypair {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		slog.Error("Failed to generate an ssh keypair.", slog.Any("propagatedError", err));
		os.Exit(42);
	}

	privBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		slog.Error("Failed to marshal private key", slog.Any("propagatedError", err));
		os.Exit(42);
	}

	privPem := pem.EncodeToMemory(privBlock)

	sshPubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		slog.Error("Failed deriving public ssh key from a private one.", slog.Any("propagatedError", err));
		os.Exit(42);
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPubKey)

	return Keypair{
		Private: string(privPem),
		Public:  string(pubBytes),
	}
}
