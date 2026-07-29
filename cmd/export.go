package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var exportFormat string
var exportFile string

type exportRecord struct {
	Service  string `json:"service"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Note     string `json:"note,omitempty"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export vault to a file",
	Long:  `Export all vault records to a JSON file. The exported file will contain passwords in plaintext.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		fmt.Println()
		common.YellowPrintln("⚠️ WARNING: You are about to export ALL passwords in PLAINTEXT to a file.")
		common.YellowPrintln(" Anyone with access to this file can read your passwords.")
		fmt.Print("Type 'Export' to confirm: ")

		confirmation, err := readLine()
		if err != nil {
			return err
		}
		if confirmation != "Export" {
			common.RedPrintln("Export cancelled.")
			return nil
		}

		records := vaultService.Records()
		if len(records) == 0 {
			common.YellowPrintln("Vault is empty. Nothing to export.")
			return nil
		}

		var exportData []exportRecord
		for _, rec := range records {
			exportData = append(exportData, exportRecord{
				Service:  string(rec.Service),
				Login:    string(rec.Login),
				Password: string(rec.Password),
				Note:     string(rec.Note),
			})
		}

		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		defer common.ZeroBytes(jsonData)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}

		err = os.WriteFile(exportFile, jsonData, 0600)
		if err != nil {
			return fmt.Errorf("failed to write export file: %w", err)
		}

		fmt.Println()
		common.GreenPrintln("Export successful!")
		fmt.Printf("Data saved to: %s\n", exportFile)
		common.YellowPrintln("Remember to securely delete this file (e.g., 'shred' or 'srm') when done.")
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFile, "file", "f", "upass_export.json", "Output file path")
	rootCmd.AddCommand(exportCmd)
}
