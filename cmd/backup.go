package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage vault backups",
	Long:  `Create, list, and restore vault backups to protect against accidental data loss.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long:  `Display all backups for the current vault, sorted by date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := store.DefaultBackupConfig()
		backups, err := store.ListBackups(vaultService.Filename(), config)
		if err != nil {
			return fmt.Errorf("list backups: %w", err)
		}

		if len(backups) == 0 {
			common.YellowPrintln("No backups found")
			fmt.Println("Backups are created automatically when the vault is saved.")
			fmt.Printf("Backup directory: %s\n", config.Directory)
			return nil
		}

		common.CyanPrintln("Available backups:")
		fmt.Printf("Directory: %s\n\n", config.Directory)
		for i, b := range backups {
			fmt.Printf("%d. %s\n", i+1, filepath.Base(b.Path))
			fmt.Printf("   Size: %d bytes | Date: %s\n", b.Size, b.ModTime.Format("2006-01-02 15:04:05"))
		}

		fmt.Println()
		backupWord := "backups"
		if len(backups) == 1 {
			backupWord = "backup"
		}
		fmt.Printf("Total: %d %s (max: %d)\n", len(backups), backupWord, config.MaxBackups)
		fmt.Println("Restore with: upass backup restore <number>")
		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-number>",
	Short: "Restore vault from a backup",
	Long: `Restore the vault from a specific backup.
A pre-restore backup of the current vault will be saved automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		num, err := strconv.Atoi(args[0])
		if err != nil || num < 1 {
			return fmt.Errorf("backup number must be a positive integer, got: %s", args[0])
		}

		config := store.DefaultBackupConfig()
		backups, err := store.ListBackups(vaultService.Filename(), config)
		if err != nil {
			return fmt.Errorf("list backups: %w", err)
		}

		if len(backups) == 0 {
			return fmt.Errorf("no backups available")
		}

		if num > len(backups) {
			return fmt.Errorf("backup number %d out of range (1-%d)", num, len(backups))
		}

		backup := backups[num-1]

		fmt.Println()
		common.YellowPrintln("⚠️  WARNING: This will permanently overwrite your current vault!")
		common.YellowPrintln("    Any records added after this backup's date will be LOST.")
		fmt.Println()
		fmt.Printf("Backup to restore: %s\n", filepath.Base(backup.Path))
		fmt.Printf("Backup date: %s\n", backup.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Println("A pre-restore backup of the current vault will be saved automatically.")
		fmt.Print("Continue? (y/n): ")

		if !readConfirmation() {
			common.RedPrintln("Cancelled")
			return nil
		}

		if err := store.RestoreBackup(backup.Path, vaultService.Filename()); err != nil {
			return fmt.Errorf("restore backup: %w", err)
		}

		if err := vaultService.Unlock(password); err != nil {
			return fmt.Errorf("vault restored on disk, but failed to unlock it with current password: %w. You may need to use 'upass recover'", err)
		}

		saveServicesCache(vaultService.ListServices())

		common.GreenPrintln("Vault restored and reloaded successfully!")
		return nil
	},
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup manually",
	Long:  `Create an immediate backup of the current vault. Backups are also created automatically on save.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := store.DefaultBackupConfig()

		if err := store.CreateBackup(vaultService.Filename(), config); err != nil {
			return fmt.Errorf("create backup: %w", err)
		}

		common.GreenPrintln("Backup created successfully!")
		fmt.Printf("Location: %s\n", config.Directory)
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupCreateCmd)
	rootCmd.AddCommand(backupCmd)
}
