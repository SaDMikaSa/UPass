package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = "config.json"
	VaultsDirName  = "vaults"
)

type AppConfig struct {
	ActiveVault string `json:"active_vault"`
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".upass", ConfigFileName)
}

func GetVaultsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".upass", "vaults")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

func GetVaultPathByName(name string) string {
	return filepath.Join(GetVaultsDir(), name+".json")
}

func LoadConfig() (*AppConfig, error) {
	path := GetConfigPath()
	if path == "" {
		return &AppConfig{ActiveVault: "default"}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{ActiveVault: "default"}, nil
		}
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &AppConfig{ActiveVault: "default"}, nil
	}

	if cfg.ActiveVault == "" {
		cfg.ActiveVault = "default"
	}

	return &cfg, nil
}

func SaveConfig(cfg *AppConfig) error {
	path := GetConfigPath()
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
