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
	Long: `Irreversibly delete the vault file, all backups, the service cache, 
		and attempt to remove the upass binary from your local bin directory.
		After this, you can run 'upass init' to start fresh (if you reinstall the binary).

		This action CANNOT be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath := vaultService.Filename()

		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			fmt.Println(common.Yellow("No vault found. Nothing to destroy."))
			return nil
		}

		// 1. Подтверждение владением (master password)
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		// 2. Показываем, что будет удалено
		fmt.Println()
		fmt.Println(common.Red("The following will be PERMANENTLY DELETED:"))
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

		// Добавляем информацию о бинарнике
		binPath := getExpectedBinPath(home)
		if _, err := os.Stat(binPath); err == nil {
			fmt.Printf("  • Binary:      %s\n", binPath)
		}

		fmt.Println()
		fmt.Println(common.Red("⚠️  This action CANNOT be undone. All passwords will be lost."))
		fmt.Println()

		// 3. Подтверждение полным словом
		fmt.Print("Type 'destroy' to confirm: ")
		confirmation, err := readLine()
		if err != nil {
			return err
		}
		if confirmation != "destroy" {
			fmt.Println(common.Red("Cancelled. Vault is intact."))
			return nil
		}

		// 4. Удаляем vault (безопасная перезапись)
		if err := secureRemove(vaultPath); err != nil {
			return fmt.Errorf("delete vault: %w", err)
		}

		// 5. Удаляем бэкапы
		for _, b := range backups {
			if err := os.Remove(b.Path); err != nil {
				fmt.Println(common.Yellow("Warning: could not remove backup %s: %v", b.Path, err))
			}
		}

		// 6. Удаляем pre-restore backup и lock-файл
		os.Remove(vaultPath + ".pre-restore.bak")
		os.Remove(vaultPath + ".lock")

		// 7. Удаляем кэш сервисов
		os.Remove(cachePath)

		// 8. Пытаемся удалить бинарный файл
		removeInstalledBinary(home)

		fmt.Println()
		fmt.Println(common.Green("Vault and all associated data have been permanently deleted."))
		fmt.Println(common.Cyan("If you wish to use UPass again, you will need to reinstall the binary and run 'upass init'."))
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

// removeInstalledBinary attempts to delete the upass binary from the standard
// installation directories. It prints a warning if the OS prevents deletion
// (e.g., because the file is currently running).
func removeInstalledBinary(home string) {
	binPath := getExpectedBinPath(home)

	// Также проверим ~/bin/upass на всякий случай (некоторые пользователи ставят туда)
	altBinPath := filepath.Join(home, "bin", "upass")
	pathsToCheck := []string{binPath}
	if binPath != altBinPath {
		pathsToCheck = append(pathsToCheck, altBinPath)
	}

	removedAny := false
	for _, path := range pathsToCheck {
		if _, err := os.Stat(path); err == nil {
			err := os.Remove(path)
			if err != nil {
				fmt.Println(common.Yellow("⚠️  Warning: Could not remove binary at %s", path))
				fmt.Println(common.Yellow("   (The file may be in use. Please delete it manually.)"))
			} else {
				fmt.Println(common.Green("Removed installed binary from: %s", path))
				removedAny = true
			}
		}
	}

	if !removedAny {
		_, err := os.Executable()
		if err == nil {
			fmt.Println(common.Yellow("ℹ️  Note: The upass binary was not found in standard installation directories."))
			fmt.Println(common.Yellow("   If you are running a local build (e.g., ./upass), you may want to delete it manually."))
		}
	}
}

// secureRemove overwrites the file with random data then zeros before deletion.
func secureRemove(path string) error {
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
		f.WriteAt(randomData, 0)
		f.Sync()
	}

	zeroData := make([]byte, size)
	f.WriteAt(zeroData, 0)
	f.Sync()

	f.Close()
	return os.Remove(path)
}
