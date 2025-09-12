package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/color"
	"github.com/SaDMikaSa/UPass/internal/storage"
)

var (
	reader         = bufio.NewReader(os.Stdin)
	storageManager *storage.StorageManager
)

// InitStorage initializes the storage manager with the default database file.
func InitStorage() error {
	var err error
	storageManager, err = storage.NewStorageManager(".src.gotlib")
	return err
}

/*
createRecord prompts user for service, login, and password details,
then stores the record in the database.
*/
func createRecord() error {
	service, err := readInput("service name")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	login, err := readInput("username/email/login")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	password, err := readInput("password")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	if err := storageManager.AddPassword(service, login, password); err != nil {
		return fmt.Errorf("failed to save record: %w", err)
	}

	color.PrintSuccess("Record saved successfully!")

	return nil
}

// getRecord retrieves and displays password record by service name.
func getRecord() error {
	service, err := readInput("service name to search")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	record, err := storageManager.GetPassword(service)
	if err != nil {
		color.PrintRejected("Record not found")
		return nil
	}

	color.PrintInfo(fmt.Sprintf("Username: %s", record.Username))
	color.PrintInfo(fmt.Sprintf("Password: %s", record.Password))
	return nil
}

// listRecords displays all stored password records with pagination.
func addRecord() error {
	color.PrintInfo("record add functionality coming soon...")
	return nil
}

// changeRecord updates an existing password record.
// Allows modification of username and/or password for specified service.
func changeRecord() error {
	service, err := readInput("service name to change")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	record, err := storageManager.GetPassword(service)
	if err != nil {
		color.PrintRejected("Service not found")
		return nil
	}

	color.PrintInfo(fmt.Sprintf("Current username: %s", record.Username))
	newUsername, err := readInput("new username (press enter to keep current)")
	if err == nil && newUsername != "" {
		record.Username = newUsername
	}
	newPassword, err := readInput("new password (press enter to keep current)")
	if err == nil && newPassword != "" {
		record.Password = newPassword
	}

	if err := storageManager.AddPassword(record.Service, record.Username, record.Password); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	color.PrintSuccess("Record updated successfully!")

	return nil
}

// deleteRecord removes a password record after user confirmation.
func deleteRecord() error {
	service, err := readInput("service name to delete")
	if err != nil {
		color.PrintRejected("Invalid input")
		return nil
	}

	record, err := storageManager.GetPassword(service)
	if err != nil {
		color.PrintRejected("Service not found")
		return nil
	}
	color.PrintInfo(fmt.Sprintf("Are you sure you want to delete [%s]?", record.Service))
	color.PrintPrompts("[yes] or [no]")
	confirm, _ := reader.ReadString('\n')

	switch strings.TrimSpace(strings.ToLower(confirm)) {
	case "y", "yes", "нуы", "н", "д":
		if err := storageManager.DeletePassword(service); err != nil {
			return fmt.Errorf("failed to delete: %w", err)
		}
		color.PrintSuccess("Record deleted successfully!")
		return nil
	default:
		color.PrintInfo("Deletion cancelled")
		return nil
	}
}

// exit terminates the application.
func exit() error {
	color.PrintInfo("App is closed")
	os.Exit(0)
	return nil
}

func changeUPass() error {
	color.PrintInfo("{{UPass}} change functionality coming soon...")
	return nil
}

// readInput reads and validates user input for the given prompt.
func readInput(prompt string) (string, error) {
	color.PrintPrompts(prompt)
	data, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	data = strings.TrimSpace(data)
	if data == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	return data, nil
}
