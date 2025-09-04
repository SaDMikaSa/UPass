package models

// Config represents the configuration settings for the application.
type Config struct {
	IsFirstRun bool   `json:"is_first_run"`
	Password   string `json:"password"`
	Word       string `json:"word"`
	Power      int    `json:"power"`
}

// defaultsConfig returns a new Config instance with default values.
func DefaultsConfig() *Config {
	return &Config{
		IsFirstRun: true,
		Password:   "",
		Word:       "",
		Power:      1,
	}
}
