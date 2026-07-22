package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var editCmd = &cobra.Command{
	Use:   "edit <service>",
	Short: "Edit an existing record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		oldRecord, err := vaultService.GetRecord(args[0])
		if err != nil {
			return fmt.Errorf("get record: %w", err)
		}
		defer common.ZeroBytes(oldRecord.Service)
		defer common.ZeroBytes(oldRecord.Login)
		defer common.ZeroBytes(oldRecord.Note)
		defer common.ZeroBytes(oldRecord.Password)

		newRecord, err := editService(oldRecord)
		if err != nil {
			return err
		}
		defer common.ZeroBytes(newRecord.Service)
		defer common.ZeroBytes(newRecord.Login)
		defer common.ZeroBytes(newRecord.Note)
		defer common.ZeroBytes(newRecord.Password)

		if err := vaultService.EditRecord(args[0], newRecord, password); err != nil {
			return fmt.Errorf("edit record: %w", err)
		}

		saveServicesCache(vaultService.ListServices())

		fmt.Println(common.Green("Record updated successfully"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.ValidArgsFunction = serviceCompletion
}

// editService prompts the user to modify fields of oldRecord. Leaving inputs
// empty keeps the current value for that field.
func editService(oldRecord domain.Record) (domain.Record, error) {
	record := domain.Record{}

	fmt.Printf("Service. Current value: %s (leave empty to keep current): ", string(oldRecord.Service))
	service, err := readLine()
	if err != nil {
		return record, err
	}

	var serviceBytes []byte
	if service == "" {
		serviceBytes = make([]byte, len(oldRecord.Service))
		copy(serviceBytes, oldRecord.Service)
	} else {
		serviceBytes = []byte(service)
	}

	fmt.Printf("Login. Current value: %s (leave empty to keep current): ",
		oldRecord.Login)
	login, err := readLine()
	if err != nil {
		return record, err
	}

	var loginBytes []byte
	if login == "" {
		loginBytes = make([]byte, len(oldRecord.Login))
		copy(loginBytes, oldRecord.Login)
	} else {
		loginBytes = []byte(login)
	}

	fmt.Print("New password (leave empty to keep current): ")
	password, err := inputOptionalPassword()
	if err != nil {
		return record, err
	}
	defer common.ZeroBytes(password)

	if len(password) == 0 {
		record.Password = make([]byte, len(oldRecord.Password))
		copy(record.Password, oldRecord.Password)
	} else {
		var passwordAgain []byte

		fmt.Print("Confirm the password: ")
		passwordAgain, err = inputOptionalPassword()
		defer common.ZeroBytes(passwordAgain)
		if err != nil {
			return record, err
		}

		if !bytes.Equal(password, passwordAgain) {
			return record, common.ErrNotMatched
		}

		record.Password = make([]byte, len(password))
		copy(record.Password, password)
	}

	fmt.Printf("Note. Current value: %s (leave empty to keep current): ", string(oldRecord.Note))
	note, err := readLine()
	if err != nil {
		return record, err
	}

	if note == "" {
		record.Note = make([]byte, len(oldRecord.Note))
		copy(record.Note, oldRecord.Note)
	} else {
		record.Note = []byte(note)
	}

	record.Service = serviceBytes
	record.Login = make([]byte, len(loginBytes))
	copy(record.Login, loginBytes)

	if note != "" {
		fmt.Println(common.Cyan("| Service: %s | Login: %s | Note: %s |", service, login, note))
	} else {
		fmt.Println(common.Cyan("| Service: %s | Login: %s |", service, login))
	}

	fmt.Print("Confirm? (y/n): ")
	if !readConfirmation() {
		return record, common.ErrCanceled
	}

	return record, nil
}

// inputOptionalPassword reads a password without requiring a non-empty value.
func inputOptionalPassword() ([]byte, error) {
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("error reading password: %w", err)
	}

	fmt.Println()
	return pass, nil
}
