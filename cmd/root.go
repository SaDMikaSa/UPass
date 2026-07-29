package cmd

import (
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/service"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/spf13/cobra"
)

var (
	vaultService *service.VaultService
)

func init() {
	vaultPath, err := store.DefaultVaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: failed to initialize vault directory: %v\n", err)
		os.Exit(1)
	}

	vaultService = service.NewVaultService(vaultPath)
}

var rootCmd = &cobra.Command{
	Use:   "upass",
	Short: "CLI password manager",
	Long:  `UPass - local encrypted password manager. Securely store, add, and retrieve credentials.`,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root cobra command and sets up completion options.
// It exits the process with a non-zero code when command execution fails.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.CompletionOptions.DisableDescriptions = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
