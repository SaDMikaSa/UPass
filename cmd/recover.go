package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover vault access and set new master password",
	Long:  "Use your recovery key to unlock the vault and set a new master password.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter recovery key: ")
		recoveryKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("read recovery key: %w", err)
		}
		fmt.Println()
		defer common.ZeroBytes(recoveryKeyBytes)

		if len(recoveryKeyBytes) == 0 {
			return common.ErrNotEmpty
		}
		oldMasterPassword, err := vaultService.RecoverVault(string(recoveryKeyBytes))
		if err != nil {
			return fmt.Errorf("recovery failed: %w", err)
		}
		defer common.ZeroBytes(oldMasterPassword)

		if err := vaultService.Unlock(oldMasterPassword); err != nil {
			return fmt.Errorf("unlock failed: %w", err)
		}

		fmt.Println(common.Green("Key accepted. Set a new master password."))
		newPassword, err := inputPass("New master password: ")
		if err != nil {
			return err
		}
		defer common.ZeroBytes(newPassword)

		strength := zxcvbn.PasswordStrength(string(newPassword), nil)
		if strength.Score < common.MinStrengthScore {
			fmt.Println(common.Yellow("⚠️  Weak password (score: %d/4, crack time: %s)", strength.Score, strength.CrackTimeDisplay))
			fmt.Print("Continue anyway? (y/n): ")
			if !readConfirmation() {
				fmt.Println(common.Red("Cancelled"))
				return nil
			}
		}

		newPasswordAgain, err := inputPass("Confirm new master password: ")
		if err != nil {
			return err
		}
		defer common.ZeroBytes(newPasswordAgain)

		if !bytes.Equal(newPassword, newPasswordAgain) {
			return common.ErrNotMatched
		}

		if err := vaultService.ChangePassword(oldMasterPassword, newPassword, string(recoveryKeyBytes)); err != nil {
			return fmt.Errorf("change password: %w", err)
		}

		fmt.Println(common.Green("Master password changed successfully!"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}
