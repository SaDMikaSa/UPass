package cmd

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/spf13/cobra"
)

var addGenerate bool

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new record",
	Long: `Add a new password record to the vault.
		For multiple accounts on the same service, use 'service:tag':
  		upass add github:work
  		upass add github:personal`,

	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		record, err := inputService(addGenerate, args...)
		if err != nil {
			if errors.Is(err, common.ErrCanceled) {
				fmt.Println(common.Red("Cancelled"))
				return nil
			}
			return err
		}
		defer common.ZeroBytes(record.Service)
		defer common.ZeroBytes(record.Login)
		defer common.ZeroBytes(record.Note)
		defer common.ZeroBytes(record.Password)

		if err := vaultService.AddRecord(record, password); err != nil {
			if errors.Is(err, common.ErrDuplicate) {
				fmt.Println(common.Yellow("Service already exists."))
				fmt.Println(common.Yellow("Use 'service:tag' for multiple accounts, e.g.:"))
				fmt.Printf("  upass add %s:work\n", record.Service)
			}

			return err
		}

		saveServicesCache(vaultService.ListServices())

		fmt.Println(common.Green("Record added successfully"))
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&addGenerate, "generate", "g", false, "Generate password")

	rootCmd.AddCommand(addCmd)
}

// inputService collects service/login/password/note interactively from the user.
// If the service arg is provided, it skips prompting for the service name.
func inputService(generate bool, args ...string) (domain.Record, error) {
	record := domain.Record{}
	var service string

	if len(args) > 0 && args[0] != "" {
		service = args[0]
		fmt.Printf("Service: %s\n", service)
	} else {
		fmt.Print("Service: ")
		var err error
		service, err = readLine()
		if err != nil {
			return record, err
		}
	}
	if service == "" {
		return record, common.ErrNotEmpty
	}

	fmt.Print("Login: ")
	login, err := readLine()
	if err != nil {
		return record, err
	}

	if login == "" {
		return record, common.ErrNotEmpty
	}

	var password []byte
	if generate {
		password = interactiveGenerate()
		if password == nil {
			return record, fmt.Errorf("password generation cancelled")
		}
	} else {
		password, err = inputPass("Input password: ")
		if err != nil {
			return record, err
		}

		passwordAgain, err := inputPass("Confirm the password: ")
		if err != nil {
			common.ZeroBytes(password)
			return record, err
		}
		defer common.ZeroBytes(passwordAgain)

		if !bytes.Equal(password, passwordAgain) {
			return record, common.ErrNotMatched
		}
	}
	defer common.ZeroBytes(password)

	fmt.Print("Note (optional): ")
	note, err := readLine()
	if err != nil {
		return record, err
	}

	if note != "" {
		fmt.Printf("| Service: %s | Login: %s | Note: %s |\n", service, login, note)
	} else {
		fmt.Printf("| Service: %s | Login: %s |\n", service, login)
	}

	fmt.Print("Confirm? (y/n): ")
	if !readConfirmation() {
		return record, common.ErrCanceled
	}

	record.Service = []byte(service)
	record.Login = []byte(login)
	if note != "" {
		record.Note = []byte(note)
	}
	record.Password = make([]byte, len(password))
	copy(record.Password, password)

	return record, nil
}
