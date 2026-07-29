package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
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
		defer common.ZeroBytes(recoveryKeyBytes)
		if err != nil {
			return fmt.Errorf("read recovery key: %w", err)
		}
		fmt.Println()

		if len(recoveryKeyBytes) == 0 {
			return common.ErrNotEmpty
		}
		oldMasterPassword, err := vaultService.RecoverVault(recoveryKeyBytes)
		defer common.ZeroBytes(oldMasterPassword)
		if err != nil {
			return fmt.Errorf("recovery failed: %w", err)
		}

		if err := vaultService.Unlock(oldMasterPassword); err != nil {
			return fmt.Errorf("unlock failed: %w", err)
		}

		common.GreenPrintln("Key accepted. Set a new master password.")
		newPassword, err := newMasterPassword()
		defer common.ZeroBytes(newPassword)
		if err != nil {
			return err
		}

		newPasswordAgain, err := inputPass("Confirm new master password: ")
		defer common.ZeroBytes(newPasswordAgain)
		if err != nil {
			return err
		}

		if !bytes.Equal(newPassword, newPasswordAgain) {
			return common.ErrNotMatched
		}

		if err := vaultService.ChangePassword(oldMasterPassword, newPassword, recoveryKeyBytes); err != nil {
			return fmt.Errorf("change password: %w", err)
		}

		common.GreenPrintln("Master password changed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}
