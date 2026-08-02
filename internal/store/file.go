package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/crypto"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/gofrs/flock"
)

func Save(filename string, vault domain.Vault, password []byte) error {
	lockPath := filename + ".lock"
	lock := flock.New(lockPath)

	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return common.ErrVaultLockedByAnother
	}
	defer lock.Unlock()

	if _, err := os.Stat(filename); err == nil {
		config := DefaultBackupConfig()
		if backupErr := CreateBackup(filename, config); backupErr != nil {
			return fmt.Errorf("failed to create pre-write backup: %w", backupErr)
		}
	}

	data, err := marshalVaultBinary(vault.Records)
	if err != nil {
		return err
	}
	defer common.ZeroBytes(data)

	encrypted, err := crypto.Encrypt(data, password, vault.EncryptedMasterPass)
	if err != nil {
		return err
	}

	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, encrypted, 0600); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if f, err := os.OpenFile(tmpFile, os.O_WRONLY, 0); err == nil {
		_ = f.Sync()
		f.Close()
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename temp file: %w", err)
	}

	if dir, err := os.Open(filepath.Dir(filename)); err == nil {
		_ = dir.Sync()
		dir.Close()
	}

	return nil
}

func Load(filename string, password []byte) (domain.Vault, error) {
	var vault domain.Vault

	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return vault, common.ErrNotFound
		}
		return vault, common.ErrReadVaultFile
	}

	if info.Mode().Perm()&0077 != 0 {
		return vault, fmt.Errorf("FATAL: vault file has insecure permissions: %o (expected 0600). "+
			"Other users on this system might be able to read it. "+
			"Please fix permissions manually: chmod 600 %s", info.Mode().Perm(), filename)
	}

	lockPath := filename + ".lock"
	lock := flock.New(lockPath)

	locked, err := lock.TryRLock()
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

	vault.Records, err = unmarshalVaultBinary(data)
	if err != nil {
		return vault, err
	}

	vault.EncryptedMasterPass = emp

	return vault, nil
}
