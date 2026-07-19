package cmd

import (
	"bytes"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/spf13/cobra"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change the master password",
	Long: `Change the master password of the vault.
		You will be asked whether you have the current recovery key:
  			- If yes, the key is reused and recovery keeps working.
  			- If no, recovery will be permanently disabled for this vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentPassword, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(currentPassword)

		fmt.Println()
		fmt.Println(common.Cyan("Do you have your current recovery key? (y/n)"))
		fmt.Println("  y — recovery will keep working (key will be reused)")
		fmt.Println("  n — recovery will be permanently disabled")
		fmt.Print("Answer: ")

		hasKey := readConfirmation()

		var recoveryKeyStr string
		if hasKey {
			fmt.Print("Enter your current recovery key: ")
			recoveryKeyStr, err = readLine()
			if err != nil {
				return err
			}
			if recoveryKeyStr == "" {
				return common.ErrNotEmpty
			}
		} else {
			fmt.Println()
			fmt.Println(common.Yellow("⚠️ WARNING: Recovery will be permanently disabled!"))
			fmt.Println(common.Yellow("If you forget the new master password, you will LOSE ACCESS to the vault forever."))
			fmt.Print("Are you absolutely sure? (y/n): ")
			if !readConfirmation() {
				fmt.Println(common.Red("Cancelled"))
				return nil
			}
		}

		fmt.Println()
		newPassword, err := inputPass("New master password: ")
		if err != nil {
			return err
		}
		defer common.ZeroBytes(newPassword)

		strength := zxcvbn.PasswordStrength(string(newPassword), nil)
		if strength.Score < common.MinStrengthScore {
			fmt.Println(common.Yellow("Weak password (score: %d/4, crack time: %s)", strength.Score, strength.CrackTimeDisplay))
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

		if hasKey {
			if err := vaultService.ChangePassword(currentPassword, newPassword, recoveryKeyStr); err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}
			fmt.Println(common.Green("Master password changed successfully!"))
			fmt.Println(common.Green("Recovery with your existing key will continue to work."))
		} else {
			if err := vaultService.DropRecoveryAndChangePassword(newPassword); err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}
			fmt.Println(common.Green("Master password changed successfully."))
			fmt.Println(common.Yellow("Recovery has been disabled."))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(passwdCmd)
}
