package cmd

import (
	"bytes"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/spf13/cobra"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change the master password",
	Long: `Change the master password of the vault. You will be asked whether you have the current recovery key: 
	- If yes, the key is reused and recovery keeps working. 
	- If no, recovery will be permanently disabled for this vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentPassword, err := unlock()
		defer common.ZeroBytes(currentPassword)
		if err != nil {
			return err
		}

		fmt.Println()
		common.CyanPrintln("Do you have your current recovery key? (y/n)")
		fmt.Println("  y — recovery will keep working (key will be reused)")
		fmt.Println("  n — recovery will be permanently disabled")
		fmt.Print("Answer: ")

		hasKey := readConfirmation()

		var recoveryKey []byte
		if hasKey {
			recoveryKey, err = inputPass("Enter your current recovery key: ")
			defer common.ZeroBytes(recoveryKey)
			if err != nil {
				return err
			}
			if recoveryKey == nil {
				return common.ErrNotEmpty
			}
		} else {
			fmt.Println()
			common.YellowPrintln("⚠️  WARNING: Recovery will be PERMANENTLY DISABLED!")
			common.YellowPrintln("If you forget the new master password, you will LOSE ACCESS to the vault FOREVER.")
			fmt.Println()

			common.YellowPrintln("Creating a mandatory pre-destruction backup...")
			config := store.DefaultBackupConfig()
			if err := store.CreateBackup(vaultService.Filename(), config); err != nil {
				return fmt.Errorf("failed to create mandatory backup: %w", err)
			}
			common.GreenPrintln("Backup created successfully.")
			fmt.Println()

			common.YellowPrintln("To proceed, you must type the word 'DESTROY' exactly as shown.")
			fmt.Print("Type 'DESTROY' to confirm disabling recovery: ")

			confirmation, err := readLine()
			if err != nil {
				return err
			}
			if confirmation != "DESTROY" {
				common.RedPrintln("Cancelled")
				return nil
			}
		}

		fmt.Println()
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

		if hasKey {
			if err := vaultService.ChangePassword(currentPassword, newPassword, recoveryKey); err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}
			common.GreenPrintln("Master password changed successfully!")
			common.GreenPrintln("Recovery with your existing key will continue to work.")
		} else {
			if err := vaultService.DropRecoveryAndChangePassword(newPassword); err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}
			common.GreenPrintln("Master password changed successfully.")
			common.YellowPrintln("Recovery has been disabled.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(passwdCmd)
}
