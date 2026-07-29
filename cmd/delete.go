package cmd

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <service>",
	Short: "Delete a record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		service := args[0]

		common.YellowPrintln("⚠️  WARNING: This action cannot be undone!")
		fmt.Printf("Are you sure you want to delete '%s'? (y/n): ", service)
		if !readConfirmation() {
			common.RedPrintln("Cancelled")
			return nil
		}

		if err := vaultService.DeleteRecord(service, password); err != nil {
			return err
		}

		saveServicesCache(vaultService.ListServices())

		common.GreenPrintln("Record deleted successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.ValidArgsFunction = serviceCompletion
}
