package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/spf13/cobra"
)

var importFile string

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import records from a JSON or CSV file",
	Long:  `Import records from a JSON or CSV file (e.g., exported from Bitwarden or UPass). Duplicates will be skipped with a warning.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		if strings.HasSuffix(strings.ToLower(importFile), ".csv") {
			return importFromCSV(importFile, password)
		}
		return importFromJSON(importFile, password)
	},
}

func importFromJSON(filepath string, password []byte) error {
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}
	defer common.ZeroBytes(fileData)

	var importData []struct {
		Service  string `json:"service"`
		Login    string `json:"login"`
		Password string `json:"password"`
		Note     string `json:"note,omitempty"`
	}

	if err := json.Unmarshal(fileData, &importData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	var records []domain.Record
	for _, item := range importData {
		records = append(records, domain.Record{
			Service:  []byte(item.Service),
			Login:    []byte(item.Login),
			Password: []byte(item.Password),
			Note:     []byte(item.Note),
		})
	}

	return processImportRecords(records, password)
}

func importFromCSV(filepath string, password []byte) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}

	requiredCols := []string{"name", "login", "password"}
	for _, col := range requiredCols {
		if _, exists := colIndex[col]; !exists {
			return fmt.Errorf("CSV is missing required column: '%s'", col)
		}
	}

	var records []domain.Record
	for {
		row, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to read CSV row: %w", err)
		}

		serviceBytes := []byte(strings.TrimSpace(row[colIndex["name"]]))
		loginBytes := []byte(strings.TrimSpace(row[colIndex["login"]]))
		passBytes := []byte(row[colIndex["password"]])

		var noteBytes []byte
		if idx, ok := colIndex["notes"]; ok && idx < len(row) {
			noteBytes = []byte(strings.TrimSpace(row[idx]))
		}

		records = append(records, domain.Record{
			Service:  serviceBytes,
			Login:    loginBytes,
			Password: passBytes,
			Note:     noteBytes,
		})
	}

	return processImportRecords(records, password)
}

func processImportRecords(records []domain.Record, password []byte) error {
	if len(records) == 0 {
		common.YellowPrintln("No records found in the import file.")
		return nil
	}

	fmt.Printf("Found %d records to import. Processing in batch...\n", len(records))

	success, skip, err := vaultService.AddRecordsBatch(records, password)
	if err != nil {
		return fmt.Errorf("batch import failed: %w", err)
	}

	saveServicesCache(vaultService.ListServices())

	fmt.Println()
	common.GreenPrintln("Import completed!")
	fmt.Printf("Successfully imported: %d\n", success)
	if skip > 0 {
		fmt.Printf("Skipped (empty or duplicates): %d\n", skip)
	}

	for i := range records {
		common.ZeroBytes(records[i].Password)
		common.ZeroBytes(records[i].Login)
		common.ZeroBytes(records[i].Note)
		common.ZeroBytes(records[i].Service)
	}

	return nil
}

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "Input file path (JSON or CSV)")
	_ = importCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(importCmd)
}
