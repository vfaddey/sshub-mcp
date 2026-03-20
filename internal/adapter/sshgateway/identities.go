package sshgateway

import (
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func sshHomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return os.UserHomeDir()
}

func sshAgentUsable() bool {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return false
	}
	st, err := os.Stat(sock)
	return err == nil && st.Mode()&os.ModeSocket != 0
}

func userKnownHostsPath() (string, error) {
	home, err := sshHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "known_hosts"), nil
}

func defaultIdentityPaths() ([]string, error) {
	home, err := sshHomeDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(home, ".ssh")
	names := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
	var out []string
	for _, n := range names {
		p := filepath.Join(base, n)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			out = append(out, p)
		}
	}
	return out, nil
}

func loadDefaultIdentitySigners() ([]ssh.Signer, error) {
	paths, err := defaultIdentityPaths()
	if err != nil {
		return nil, err
	}
	var signers []ssh.Signer
	for _, p := range paths {
		pemBytes, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			continue
		}
		signers = append(signers, s)
	}
	return signers, nil
}
