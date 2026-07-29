package store

import (
	"os"
	"path/filepath"
)

const (
	DefaultVaultDir  = ".upass"
	DefaultVaultName = "vault"
)

func DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultVaultName, nil
	}
	dir := filepath.Join(home, DefaultVaultDir)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return DefaultVaultName, err
	}

	return filepath.Join(dir, DefaultVaultName), nil
}
