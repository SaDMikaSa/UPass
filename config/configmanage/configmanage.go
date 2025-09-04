package configmanage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/SaDMikaSa/UPass/config/inputdata"
	"github.com/SaDMikaSa/UPass/config/models"
	"github.com/SaDMikaSa/UPass/config/obfuscator"
	"github.com/SaDMikaSa/UPass/internal/color"
)

const (
	configPath = ".svtea_src"
)

func printWelcomeMessages() {
	var helloMessage = []string{
		"!----You are the first time using {{UPass}}---!",
		//"!---All data must be entered in English---!",
		"!---You can create YOUR PERSONAL PASSWORD {{UPass}} only one time---!",
		"!---If you forget {{UPass}}, you will never be able to recover your data---!",
		"!---{{UPass}} must contain uppercase letters,lowercase letters,digits and special characters---!",
	}

	for _, msg := range helloMessage {
		color.PrintInfo(msg)
	}
}

// LoadConfig loads configuration from the default file path.
// If the file doesn't exist, returns default configuration with nil error.
func LoadConfig() (*models.Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return models.DefaultsConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	deobfuscated := obfuscator.ObfuscateData(data)
	var config models.Config
	if err := json.Unmarshal(deobfuscated, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// SaveConfig writes the provided configuration to the default file path.
func SaveConfig(config *models.Config) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	obfuscated := obfuscator.ObfuscateData(data)

	if err := os.WriteFile(configPath, obfuscated, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	color.PrintSuccess()

	return nil
}

// InitializeConfig loads configuration and applies first-run initialization logic.
func InitializeConfig() (*models.Config, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if config.IsFirstRun {
		printWelcomeMessages()
		password, err := inputdata.EnterPassword()
		if err != nil {
			return nil, fmt.Errorf("password setup failed: %w", err)
		}
		if password == "" {
			log.Fatal("{{UPass}} is empty, exiting...")
		}
		config.Password = password
		word, power, err := inputdata.EnterWord()
		if err != nil {
			return nil, fmt.Errorf("word setup failed: %w", err)
		}
		if word == "" {
			log.Fatal("secret word is empty, exiting...")
		}
		config.Word, config.Power = word, power

		config.IsFirstRun = false
		if err := SaveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to save config: %w", err)
		}
	}

	return config, nil
}
