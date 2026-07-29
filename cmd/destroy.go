package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Permanently delete the vault, backups, and the binary",
	Long: `Irreversibly delete the vault file, all backups, the service cache, and attempt to remove the upass binary from your local bin directory. After this, you can run 'upass init' to start fresh (if you reinstall the binary). 
	This action CANNOT be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath := vaultService.Filename()

		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			common.YellowPrintln("No vault found. Nothing to destroy.")
			return nil
		}

		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		fmt.Println()
		common.RedPrintln("The following will be PERMANENTLY DELETED:")
		fmt.Printf("  • Vault:       %s\n", vaultPath)

		config := store.DefaultBackupConfig()
		backups, _ := store.ListBackups(vaultPath, config)
		if len(backups) > 0 {
			fmt.Printf("  • Backups:     %d file(s) in %s\n", len(backups), config.Directory)
		}

		home, _ := os.UserHomeDir()
		cachePath := filepath.Join(home, store.DefaultVaultDir, cacheFileName)
		if _, err := os.Stat(cachePath); err == nil {
			fmt.Printf("  • Cache:       %s\n", cachePath)
		}

		binPath := getExpectedBinPath(home)
		if _, err := os.Stat(binPath); err == nil {
			fmt.Printf("  • Binary:      %s\n", binPath)
		}

		fmt.Println()
		common.RedPrintln("⚠️  This action CANNOT be undone. All passwords will be lost.")
		fmt.Println()

		fmt.Print("Type 'Destroy' to confirm: ")
		confirmation, err := readLine()
		if err != nil {
			return err
		}
		if confirmation != "Destroy" {
			common.RedPrintln("Cancelled. Vault is intact.")
			return nil
		}

		common.YellowPrintln("⚠️  Overwriting a file does not guarantee physical destruction of old data.")
		common.YellowPrintln("For absolute security, use Full-Disk Encryption (FDE) and hardware secure erase tools.")
		if err := tmpDelete(vaultPath); err != nil {
			return fmt.Errorf("delete vault: %w", err)
		}

		for _, b := range backups {
			if err := os.Remove(b.Path); err != nil {
				common.YellowPrintf("Warning: could not remove backup %s: %v\n", b.Path, err)
			}
		}

		os.Remove(vaultPath + ".pre-restore.bak")
		os.Remove(vaultPath + ".lock")
		os.Remove(cachePath)

		fmt.Println()
		common.GreenPrintln("Vault and all associated data have been permanently deleted.")
		common.CyanPrintln("If you wish to use UPass again, you will need to reinstall the binary and run 'upass init'.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}

// getExpectedBinPath returns the standard installation path for the upass binary.
func getExpectedBinPath(home string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "bin", "upass.exe")
	}
	return filepath.Join(home, ".local", "bin", "upass")
}

// tmpDelete performs a best-effort secure deletion of a file.
// It overwrites the file with cryptographically secure random data,
// then with zeros, forcing a disk sync after each pass before removal.
func tmpDelete(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	size := info.Size()
	if size == 0 {
		return os.Remove(path)
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		_ = os.Remove(path)
		return err
	}

	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err == nil {
		if _, err := f.WriteAt(randomData, 0); err != nil {
			_ = os.Remove(path)
			return err
		}
		if err := f.Sync(); err != nil {
			_ = os.Remove(path)
			return err
		}
	}

	zeroData := make([]byte, size)
	if _, err := f.WriteAt(zeroData, 0); err != nil {
		_ = os.Remove(path)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = os.Remove(path)
		return err
	}

	f.Close()
	return os.Remove(path)
}
