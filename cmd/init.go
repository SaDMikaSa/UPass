package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long: `Create a new encrypted vault with a recovery key.
		The recovery key is displayed as a QR code and temporarily saved to a file.
		The file is overwritten and removed after 60 seconds.
		Scan the QR code or copy the file before it's removed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := inputPass("Enter master password: ")
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		strength := zxcvbn.PasswordStrength(string(password), nil)
		if strength.Score < common.MinStrengthScore {
			fmt.Println(common.Yellow("⚠️  Weak password (score: %d/4, crack time: %s)", strength.Score, strength.CrackTimeDisplay))
			fmt.Print("Continue anyway? [y/n]: ")
			if !readConfirmation() {
				fmt.Println(common.Red("Cancelled"))
				return nil
			}
		}

		passwordAgain, err := inputPass("Confirm master password: ")
		if err != nil {
			return err
		}
		defer common.ZeroBytes(passwordAgain)

		if !bytes.Equal(password, passwordAgain) {
			return common.ErrNotMatched
		}

		recoveryKey, err := vaultService.InitWithRecovery(password)
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}

		tmpFile, err := saveRecoveryKeyTemp(recoveryKey)
		if err != nil {
			fmt.Println(common.Red("Failed to create temp file: %v", err))
			showRecoveryQR(recoveryKey)
			fmt.Println(common.Yellow("Recovery key: %s", recoveryKey))
			return nil
		}

		showRecoveryQR(recoveryKey)

		fmt.Println(common.Green("Vault created successfully!"))
		fmt.Println(common.Yellow("⚠️ RECOVERY KEY — SAVE IT NOW:"))
		fmt.Printf("	Scan the QR code above with your phone, or\n")
		fmt.Printf("	Copy the file before it's deleted:\n")
		fmt.Printf("%s\n", common.Green(tmpFile))
		fmt.Println()
		fmt.Printf("File will be securely deleted in %d seconds.\n", 60)
		fmt.Println(common.Yellow("Press Enter to delete the file now, or wait 60 seconds for automatic deletion."))
		fmt.Println(common.Yellow("Make sure you have saved the recovery key before pressing Enter!"))
		fmt.Println(common.Yellow("Anyone with this key can access your vault and all backups."))
		fmt.Println(common.Yellow("Store it offline: paper, another device, encrypted USB."))
		fmt.Println("To recover: upass recover <key>")

		done := make(chan bool)
		go func() {
			time.Sleep(60 * time.Second)
			done <- true
		}()
		go func() {
			readLine()
			done <- true
		}()
		<-done

		if err := tmpDelete(tmpFile); err != nil {
			fmt.Println(common.Yellow("⚠️ WARNINg: Could not securely remove temp file: %v", err))
			fmt.Println(common.Yellow("Please delete it manually: %s", tmpFile))
		} else {
			fmt.Println(common.Green("Recovery key file removed from the filesystem."))
		}

		fmt.Println(common.Yellow("⚠️ NOTE: On modern SSDs, hardware wear-leveling may prevent guaranteed physical overwriting."))
		fmt.Println(common.Yellow("For absolute security, physical destruction of the drive is required if compromised."))

		return nil
	},
}

func saveRecoveryKeyTemp(key string) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("upass-recovery-key-%d.txt", time.Now().Unix()))

	if err := os.WriteFile(tmpFile, []byte(key), 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

func showRecoveryQR(key string) {
	qr, err := qrcode.New(key, qrcode.Medium)
	if err != nil {
		fmt.Println(common.Red("Failed to generate QR code."))
		return
	}

	fmt.Println()
	fmt.Println(qr.ToSmallString(false))
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
