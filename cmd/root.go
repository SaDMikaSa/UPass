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
	vaultPathFlag string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultPathFlag, "vault", "", "Vault name or path (e.g., 'work' or '/tmp/v.json'). Automatically switches active context.")
}

var rootCmd = &cobra.Command{
	Use:   "upass",
	Short: "CLI password manager",
	Long:  `UPass - local encrypted password manager. Securely store, add, and retrieve credentials.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if vaultService != nil {
			vaultService.Close()
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		if vaultPathFlag != "" {
			if strings.Contains(vaultPathFlag, ".json") || strings.Contains(vaultPathFlag, "/") || strings.Contains(vaultPathFlag, "\\") {
				common.GreenPrintf("Using vault: %s\n", vaultPathFlag)
			} else {
				common.GreenPrintf("Switched to vault: %s\n", vaultPathFlag)
			}
			return nil
		}

		return cmd.Help()
	},
}

func resolveVaultPath() (string, error) {
	targetName := "default"

	if vaultPathFlag != "" {
		if strings.Contains(vaultPathFlag, ".json") || strings.Contains(vaultPathFlag, "/") || strings.Contains(vaultPathFlag, "\\") {
			return expandPath(vaultPathFlag), nil
		}
		targetName = vaultPathFlag
	} else {
		cfg, err := config.LoadConfig()
		if err == nil && cfg.ActiveVault != "" {
			targetName = cfg.ActiveVault
		}
	}

	vaultsDir := config.GetVaultsDir()
	if err := os.MkdirAll(vaultsDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create vaults directory: %w", err)
	}

	if vaultPathFlag != "" && !strings.Contains(vaultPathFlag, ".json") && !strings.Contains(vaultPathFlag, "/") {
		cfg, _ := config.LoadConfig()
		if cfg.ActiveVault != targetName {
			cfg.ActiveVault = targetName
			if err := config.SaveConfig(cfg); err != nil {
				return "", fmt.Errorf("failed to save active vault configuration: %w", err)
			}
		}
	} else if vaultPathFlag == "" {
		cfg, _ := config.LoadConfig()
		if cfg.ActiveVault == "" {
			cfg.ActiveVault = "default"
			if err := config.SaveConfig(cfg); err != nil {
				return "", fmt.Errorf("failed to save default vault configuration: %w", err)
			}
		}
	}

	return config.GetVaultPathByName(targetName), nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.CompletionOptions.DisableDescriptions = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
