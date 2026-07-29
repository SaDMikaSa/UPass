package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/spf13/cobra"
)

var importFile string

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import records from a JSON file",
	Long: `Import records from a JSON file (e.g., exported from UPass or another manager).
	Duplicates will be skipped with a warning.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(importFile)
		defer common.ZeroBytes(fileData)
		if err != nil {
			return fmt.Errorf("failed to read import file: %w", err)
		}

		var importData []exportRecord
		if err := json.Unmarshal(fileData, &importData); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		if len(importData) == 0 {
			common.YellowPrintln("No records found in the import file.")
			return nil
		}

		fmt.Printf("Found %d records to import. Processing...\n", len(importData))

		successCount := 0
		skipCount := 0

		for _, item := range importData {
			record := domain.Record{
				Service:  []byte(item.Service),
				Login:    []byte(item.Login),
				Password: make([]byte, len(item.Password)),
				Note:     []byte(item.Note),
			}
			copy(record.Password, item.Password)

			if item.Service == "" || item.Login == "" {
				skipCount++
				continue
			}

			err := vaultService.AddRecord(record, password)
			common.ZeroBytes(record.Service)
			common.ZeroBytes(record.Login)
			common.ZeroBytes(record.Password)
			common.ZeroBytes(record.Note)
			if err != nil {
				if errors.Is(err, common.ErrDuplicate) {
					fmt.Printf("  [SKIP] %s (already exists)\n", item.Service)
					skipCount++
					continue
				}
				fmt.Printf("  [ERROR] %s: %v\n", item.Service, err)
				continue
			}

			successCount++
		}

		saveServicesCache(vaultService.ListServices())

		fmt.Println()
		common.GreenPrintln("Import completed!")
		fmt.Printf("Successfully imported: %d\n", successCount)
		if skipCount > 0 {
			fmt.Printf("Skipped (duplicates): %d\n", skipCount)
		}

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "upass_export.json", "Input file path")
	rootCmd.AddCommand(importCmd)
}
