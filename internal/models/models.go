package models

// Config represents the configuration settings for the application.
type Config struct {
	IsFirstRun      bool   `json:"is_first_run"`
	PasswordHash    string `json:"password_hash"`
	EncryptedWord   string `json:"encrypted_word"`
	Power           int    `json:"power"`
	EncryptionNonce string `json:"encryption_nonce"`
	EncryptionSalt  string `json:"encryption_salt"`
}

// defaultsConfig returns a new Config instance with default values.
func DefaultsConfig() *Config {
	return &Config{
		IsFirstRun:      true,
		PasswordHash:    "",
		EncryptedWord:   "",
		Power:           1,
		EncryptionNonce: "",
		EncryptionSalt:  "",
	}
}
