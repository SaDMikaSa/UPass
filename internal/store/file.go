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

func Save(filename string, vault domain.Vault, password []byte) error {
	if _, err := os.Stat(filename); err == nil {
		config := DefaultBackupConfig()
		if backupErr := CreateBackup(filename, config); backupErr != nil {
			return fmt.Errorf("failed to create pre-write backup, aborting save for safety: %w", backupErr)
		}
	}

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

	if f, openErr := os.OpenFile(tmpFile, os.O_WRONLY, 0); openErr == nil {
		_ = f.Sync()
		f.Close()
	}

	err = os.Rename(tmpFile, filename)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename temp file: %w", err)
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
	defer common.ZeroBytes(encrypted)
	if err != nil {
		if os.IsNotExist(err) {
			return vault, common.ErrNotFound
		}
		return vault, fmt.Errorf("%w: %v", common.ErrReadVaultFile, err)
	}

	data, emp, err := crypto.Decrypt(encrypted, password)
	defer common.ZeroBytes(data)
	if err != nil {
		return vault, err
	}

	err = json.Unmarshal(data, &vault)
	if err != nil {
		return vault, err
	}

	vault.EncryptedMasterPass = emp

	return vault, nil
}
