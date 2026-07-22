package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SaDMikaSa/UPass/internal/common"
)

const (
	DefaultBackupDir  = ".upass/backups"
	DefaultMaxBackups = 10
)

type BackupConfig struct {
	Enabled    bool
	Directory  string
	MaxBackups int
}

// DefaultBackupConfig returns the default backup configuration which stores
// backups under the user's home directory in .upass/backups and keeps up to
// DefaultMaxBackups entries.
func DefaultBackupConfig() BackupConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return BackupConfig{
		Enabled:    true,
		Directory:  filepath.Join(home, DefaultBackupDir), // .upass/backups
		MaxBackups: DefaultMaxBackups,
	}
}

type BackupInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// CreateBackup writes a copy of the vault file into the configured backup
// directory with a timestamp suffix, then rotates old backups according to the
// MaxBackups policy.
func CreateBackup(vaultPath string, config BackupConfig) error {
	if !config.Enabled {
		return nil
	}

	if err := os.MkdirAll(config.Directory, 0700); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	data, err := os.ReadFile(vaultPath)
	if err != nil {
		return fmt.Errorf("read vault for backup: %w", err)
	}

	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	timestamp := time.Now().Format("2006-01-02_150405")
	backupName := fmt.Sprintf("%s.%s%s", nameWithoutExt, timestamp, ext)
	backupPath := filepath.Join(config.Directory, backupName)

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	return rotateBackups(vaultPath, config)
}

// rotateBackups removes old backups when the number of backup files exceeds
// config.MaxBackups. It keeps the most recent backups based on filename sort
// order (which includes the timestamp inserted by CreateBackup).
func rotateBackups(vaultPath string, config BackupConfig) error {
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	pattern := filepath.Join(config.Directory, nameWithoutExt+".*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob backups: %w", err)
	}

	if len(matches) <= config.MaxBackups {
		return nil
	}

	sort.Strings(matches)

	toDelete := matches[:len(matches)-config.MaxBackups]
	for _, f := range toDelete {
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("failed to remove old backup %s: %w", f, err)
		}
	}

	return nil
}

// ListBackups returns metadata about available backups for the given vault
// path. It returns an ordered list of BackupInfo entries or an error if the
// backup directory cannot be enumerated.
func ListBackups(vaultPath string, config BackupConfig) ([]BackupInfo, error) {
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	pattern := filepath.Join(config.Directory, nameWithoutExt+".*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob backups: %w", err)
	}

	sort.Strings(matches)

	var backups []BackupInfo
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return backups, nil
}

// RestoreBackup replaces the active vault file with the provided backup.
// It performs atomic write with validation and creates a pre-restore backup.
func RestoreBackup(backupPath, vaultPath string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	if len(backupData) < 4 || !bytes.Equal(backupData[:4], common.VaultMagic) {
		return fmt.Errorf("backup file is not a valid UPass vault (invalid magic bytes)")
	}

	if currentData, err := os.ReadFile(vaultPath); err == nil {
		preRestorePath := vaultPath + ".pre-restore.bak"
		if err := os.WriteFile(preRestorePath, currentData, 0600); err != nil {
			return fmt.Errorf("create pre-restore backup: %w", err)
		}
	}

	tmpFile := vaultPath + ".restore.tmp"

	if err := os.WriteFile(tmpFile, backupData, 0600); err != nil {
		return fmt.Errorf("write temp restore file: %w", err)
	}

	if f, err := os.OpenFile(tmpFile, os.O_WRONLY, 0); err == nil {
		if err := f.Sync(); err != nil {
			f.Close()
			os.Remove(tmpFile)
			return fmt.Errorf("sync temp restore file: %w", err)
		}
		f.Close()
	}

	if err := os.Rename(tmpFile, vaultPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename temp restore file: %w", err)
	}

	return nil
}
