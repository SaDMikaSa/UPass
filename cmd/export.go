package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/spf13/cobra"
)

var (
	exportFile   string
	exportFormat string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export vault to a file",
	Long:  `Export all vault records to a JSON or CSV file. WARNING: The exported file will contain passwords in plaintext.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		fmt.Println()
		common.YellowPrintln("⚠️  WARNING: You are about to export ALL passwords in PLAINTEXT to a file.")
		common.YellowPrintln("   Anyone with access to this file can read your passwords.")
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

		format := exportFormat
		if format == "json" && strings.HasSuffix(strings.ToLower(exportFile), ".csv") {
			format = "csv"
		} else if format == "csv" && strings.HasSuffix(strings.ToLower(exportFile), ".json") {
			format = "json"
		}

		if format == "csv" {
			return exportToCSV(exportFile, records)
		}

		return exportToJSON(exportFile, records)
	},
}

func exportToJSON(filepath string, records map[string]domain.Record) error {
	type exportRecord struct {
		Service  string `json:"service"`
		Login    string `json:"login"`
		Password string `json:"password"`
		Note     string `json:"note,omitempty"`
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
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	defer common.ZeroBytes(jsonData)

	if err := os.WriteFile(filepath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	printExportSuccess(filepath)
	return nil
}

func exportToCSV(filepath string, records map[string]domain.Record) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"folder", "favorite", "type", "name", "notes", "fields", "reprompt", "uri", "login", "password"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, rec := range records {
		// type is always "login" for UPass
		row := []string{
			"",                   // folder
			"0",                  // favorite
			"login",              // type
			string(rec.Service),  // name (Service)
			string(rec.Note),     // notes
			"",                   // fields
			"0",                  // reprompt
			"",                   // uri
			string(rec.Login),    // login
			string(rec.Password), // password
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	printExportSuccess(filepath)
	return nil
}

func printExportSuccess(filepath string) {
	fmt.Println()
	common.GreenPrintln("Export successful!")
	fmt.Printf("Data saved to: %s\n", filepath)
	common.YellowPrintln("Remember to securely delete this file (e.g., 'shred' or 'sdelete') when done.")
}

func init() {
	exportCmd.Flags().StringVarP(&exportFile, "file", "f", "upass_export", "Output file path")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "F", "json", "Export format: 'json' or 'csv'")
	rootCmd.AddCommand(exportCmd)
}
