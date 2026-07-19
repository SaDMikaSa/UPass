package store

import (
	"os"
	"path/filepath"
)

const (
	DefaultVaultDir  = ".upass"
	DefaultVaultName = "vault"
)

// DefaultVaultPath returns the default file path used to store the vault in
// the user's home directory (~/.upass/vault). It ensures the directory exists
// with restrictive permissions.
func DefaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultVaultName
	}
	dir := filepath.Join(home, DefaultVaultDir)
	os.MkdirAll(dir, 0700)

	return filepath.Join(dir, DefaultVaultName)
}
