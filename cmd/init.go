package cmd

import (
	"bytes"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long:  `Create a new encrypted vault with a recovery key. The recovery key is displayed as a QR code and a base64 string. Save it immediately. It will not be shown again.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := newMasterPassword()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		passwordAgain, err := inputPass("Confirm master password: ")
		defer common.ZeroBytes(passwordAgain)
		if err != nil {
			return err
		}

		if !bytes.Equal(password, passwordAgain) {
			return common.ErrNotMatched
		}

		recoveryKey, err := vaultService.InitWithRecovery(password)
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}

		showRecoveryQR(recoveryKey)

		common.GreenPrintln("Vault created successfully!")
		fmt.Println()
		common.YellowPrintln("⚠️  RECOVERY KEY — SAVE IT NOW OR LOSE ACCESS FOREVER")
		common.YellowPrintln("This is the ONLY time this key will be displayed.")
		fmt.Println()
		common.YellowPrintln("Scan the QR code above with your phone, or copy the key below:")
		common.CyanPrintln(recoveryKey)
		fmt.Println()
		common.YellowPrintln("⚠️  WARNING:")
		fmt.Println("  • Store this key offline: on paper, in a password manager, or on an encrypted USB.")
		fmt.Println("  • Don't save it in a plain text file on this computer.")
		fmt.Println("  • Anyone with this key can access your vault and all backups.")
		fmt.Println()
		fmt.Println("To recover your vault in the future, run: upass recover <key>")

		return nil
	},
}

func showRecoveryQR(key string) {
	qr, err := qrcode.New(key, qrcode.Medium)
	if err != nil {
		common.RedPrintln("Failed to generate QR code.")
		return
	}
	fmt.Println()
	fmt.Println(qr.ToSmallString(false))
}

func init() {
	rootCmd.AddCommand(initCmd)
}
