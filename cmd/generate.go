package cmd

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var (
	genLength    int = 20
	genNoSymbols bool
	genExclude   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a strong random password",
	Long:  "Generate a cryptographically secure random password using crypto/rand.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if genLength < 1 {
			return fmt.Errorf("password length must be positive, got %d", genLength)
		}

		password, err := generateRandomPassword()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		fmt.Println(common.Cyan("%s", password))
		return nil
	},
}

func init() {
	generateCmd.Flags().IntVarP(&genLength, "length", "l", 20, "Password length")
	generateCmd.Flags().BoolVarP(&genNoSymbols, "no-symbols", "n", false, "Exclude special characters")
	generateCmd.Flags().StringVar(&genExclude, "exclude", "", "Characters to exclude (use quotes for spaces, e.g., --exclude \"O 0 l 1\")")
	rootCmd.AddCommand(generateCmd)
}
