package cmd

import (
	"fmt"
	"time"

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
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		record, err := vaultService.GetRecord(args[0])
		defer common.ZeroBytes(record.Service)
		defer common.ZeroBytes(record.Login)
		defer common.ZeroBytes(record.Note)
		defer common.ZeroBytes(record.Password)
		if err != nil {
			return err
		}

		if showPassword {
			common.CyanPrintf("Service: %s\nLogin: %s\nPassword: %s\n", record.Service, record.Login, record.Password)
		} else {
			if err := clipboard.WriteAll(string(record.Password)); err != nil {
				return fmt.Errorf("copy to clipboard: %w", err)
			}
			common.CyanPrintf("Service: %s\nLogin: %s\n", record.Service, record.Login)
			if len(record.Note) != 0 {
				common.CyanPrintf("Note: %s\n", record.Note)
			}
			common.GreenPrintln("Password copied to clipboard!")
			common.YellowPrintln("⚠️ WARNING: Process will stay alive for 15s to clear clipboard.")
			common.YellowPrintln("	Press Ctrl+C to exit early (clipboard will NOT be cleared).")

			for i := 15; i > 0; i-- {
				time.Sleep(1 * time.Second)
			}
			_ = clipboard.WriteAll("")
			common.GreenPrintln("Clipboard cleared. Exiting.")
		}

		return nil
	},
}

func init() {
	getCmd.Flags().BoolVarP(&showPassword, "show", "s", false, "Show password in terminal instead of copying to clipboard\nPassword will be displayed in plain text.")
	rootCmd.AddCommand(getCmd)
	getCmd.ValidArgsFunction = serviceCompletion
}
