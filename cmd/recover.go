package cmd

import (
	"bytes"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/spf13/cobra"
)

var recoverCmd = &cobra.Command{
	Use:   "recover <recovery-key>",
	Short: "Recover vault access and set new master password",
	Long:  "Use your recovery key to unlock the vault and set a new master password.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		recoveryKey := args[0]

		oldMasterPassword, err := vaultService.RecoverVault(recoveryKey)
		if err != nil {
			return fmt.Errorf("recovery failed: %w", err)
		}
		defer common.ZeroBytes(oldMasterPassword)

		if err := vaultService.Unlock(oldMasterPassword); err != nil {
			return fmt.Errorf("unlock failed: %w", err)
		}

		fmt.Println(common.Green("Key accepted."))
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

		if err := vaultService.ChangePassword(oldMasterPassword, newPassword, recoveryKey); err != nil {
			return fmt.Errorf("change password: %w", err)
		}

		fmt.Println(common.Green("Master password changed successfully!"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}
