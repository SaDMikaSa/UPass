package cmd

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var showPassword bool

var getCmd = &cobra.Command{
	Use:   "get <service>",
	Short: "Get a record by service name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		record, err := vaultService.GetRecord(args[0])
		if err != nil {
			return err
		}
		defer common.ZeroBytes(record.Service)
		defer common.ZeroBytes(record.Login)
		defer common.ZeroBytes(record.Note)
		defer common.ZeroBytes(record.Password)

		if showPassword {
			fmt.Printf("Service: %s\nLogin: %s\nPassword: %s\n", common.Cyan("%s", record.Service), common.Cyan("%s", record.Login), record.Password)
		} else {
			if err := clipboard.WriteAll(string(record.Password)); err != nil {
				return fmt.Errorf("copy to clipboard: %w", err)
			}
			fmt.Printf("Service: %s\nLogin: %s\n", common.Cyan("%s", record.Service), common.Cyan("%s", record.Login))
			fmt.Println(common.Green("Password copied to clipboard!"))
		}
		if len(record.Note) != 0 {
			fmt.Println(common.Cyan("Note: %s", record.Note))
		}

		return nil
	},
}

func init() {
	getCmd.Flags().BoolVarP(&showPassword, "show", "s", false, "Show password in terminal instead of copying to clipboard\nPassword will be displayed in plain text.")
	rootCmd.AddCommand(getCmd)
	getCmd.ValidArgsFunction = serviceCompletion
}
