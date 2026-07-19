package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/crypto"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/gofrs/flock"
)

// Save persists the given vault to disk at filename, encrypting the JSON
// payload with a key derived from password. The function takes a file lock to
// prevent concurrent writers, writes atomically via a temporary file, and
// triggers an automatic backup after a successful save.
func Save(filename string, vault domain.Vault, password []byte) error {
	lock := flock.New(filename)
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return common.ErrVaultLockedByAnother
	}
	defer lock.Unlock()

	data, err := json.Marshal(vault)
	if err != nil {
		return err
	}
	defer common.ZeroBytes(data)

	encrypted, err := crypto.Encrypt(data, password, vault.EncryptedMasterPass)
	if err != nil {
		return err
	}

	tmpFile := filename + ".tmp"
	err = os.WriteFile(tmpFile, encrypted, 0600)
	if err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	err = os.Rename(tmpFile, filename)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename temp file: %w", err)
	}

	config := DefaultBackupConfig()
	if err := CreateBackup(filename, config); err != nil {
		return fmt.Errorf("auto-backup failed: %w", err)
	}

	return nil
}

// Load reads and decrypts the vault file at filename using password and
// returns the parsed Vault structure. It obtains a shared file lock during
// the read and returns ErrNotFound when the file does not exist.
func Load(filename string, password []byte) (domain.Vault, error) {
	var vault domain.Vault

	lock := flock.New(filename)
	locked, err := lock.TryLock()
	if err != nil {
		return vault, err
	}
	if !locked {
		return vault, common.ErrVaultLockedByAnother
	}
	defer lock.Unlock()

	encrypted, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return vault, common.ErrNotFound
		}
		return vault, fmt.Errorf("%w: %v", common.ErrReadVaultFile, err)
	}
	defer common.ZeroBytes(encrypted)

	data, emp, err := crypto.Decrypt(encrypted, password)
	if err != nil {
		return vault, err
	}
	defer common.ZeroBytes(data)

	err = json.Unmarshal(data, &vault)
	if err != nil {
		return vault, err
	}

	vault.EncryptedMasterPass = emp

	return vault, nil
}
