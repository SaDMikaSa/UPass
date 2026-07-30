package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/config"
	"github.com/SaDMikaSa/UPass/internal/service"
	"github.com/spf13/cobra"
)

var (
	vaultService  *service.VaultService
	vaultPathFlag string // Глобальный флаг: принимает имя (work) или путь (/tmp/v.json)
)

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultPathFlag, "vault", "", "Vault name or path (e.g., 'work' or '/tmp/v.json'). Automatically switches active context.")
}

var rootCmd = &cobra.Command{
	Use:   "upass",
	Short: "CLI password manager",
	Long:  `UPass - local encrypted password manager. Securely store, add, and retrieve credentials.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Командам, не требующим доступа к хранилищу, инициализация не нужна
		if cmd.Name() == "completion" || (cmd.Parent() != nil && cmd.Parent().Name() == "completion") {
			return nil
		}
		if cmd.Name() == "generate" {
			return nil
		}

		vaultPath, err := resolveVaultPath()
		if err != nil {
			return fmt.Errorf("failed to resolve vault path: %w", err)
		}

		vaultService = service.NewVaultService(vaultPath)
		return nil
	},

	// Run вызывается, когда пользователь вводит "upass" без подкоманды
	RunE: func(cmd *cobra.Command, args []string) error {
		// Если передан флаг --vault без команды, переключаем контекст и выходим
		if vaultPathFlag != "" {
			// resolveVaultPath() уже обновил конфиг в PersistentPreRunE
			// Просто выводим сообщение
			if strings.Contains(vaultPathFlag, ".json") || strings.Contains(vaultPathFlag, "/") || strings.Contains(vaultPathFlag, "\\") {
				common.GreenPrintf("Using vault: %s\n", vaultPathFlag)
			} else {
				common.GreenPrintf("Switched to vault: %s\n", vaultPathFlag)
			}
			return nil
		}

		// Если флаг не переден, показываем help
		return cmd.Help()
	},
}

// resolveVaultPath определяет путь и обновляет конфиг при необходимости
func resolveVaultPath() (string, error) {
	targetName := "default"

	// 1. Если передан флаг --vault
	if vaultPathFlag != "" {
		// Если это похоже на абсолютный/относительный путь, используем его как есть
		if strings.Contains(vaultPathFlag, ".json") || strings.Contains(vaultPathFlag, "/") || strings.Contains(vaultPathFlag, "\\") {
			return expandPath(vaultPathFlag), nil
		}
		// Иначе считаем это именем хранилища (например, "work" или "default")
		targetName = vaultPathFlag
	} else {
		// 2. Если флаг не передан, берем активное хранилище из конфига
		cfg, err := config.LoadConfig()
		if err == nil && cfg.ActiveVault != "" {
			targetName = cfg.ActiveVault
		}
	}

	// Гарантируем существование директории для хранилищ
	vaultsDir := config.GetVaultsDir()
	if err := os.MkdirAll(vaultsDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create vaults directory: %w", err)
	}

	// 3. Если мы переключились на новое имя через флаг, сохраняем его в конфиг как активный
	if vaultPathFlag != "" && !strings.Contains(vaultPathFlag, ".json") && !strings.Contains(vaultPathFlag, "/") {
		cfg, _ := config.LoadConfig()
		if cfg.ActiveVault != targetName {
			cfg.ActiveVault = targetName
			_ = config.SaveConfig(cfg) // Игнорируем ошибку сохранения, чтобы не ломать работу, если диск переполнен
		}
	} else if vaultPathFlag == "" {
		// Если флаг не передан и в конфиге пусто, инициализируем дефолт
		cfg, _ := config.LoadConfig()
		if cfg.ActiveVault == "" {
			cfg.ActiveVault = "default"
			_ = config.SaveConfig(cfg)
		}
	}

	return config.GetVaultPathByName(targetName), nil
}

// expandPath раскрывает тильду (~) в домашнюю директорию
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// Execute runs the root cobra command and sets up completion options.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.CompletionOptions.DisableDescriptions = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
