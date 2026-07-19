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
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		service := args[0]

		fmt.Println(common.Yellow("⚠️ WARNING: This action cannot be undone!"))
		fmt.Printf("Are you sure you want to delete %s? (y/n): ", service)
		if !readConfirmation() {
			fmt.Println(common.Red("Cancelled"))
			return nil
		}

		if err := vaultService.DeleteRecord(service, password); err != nil {
			return err
		}

		saveServicesCache(vaultService.ListServices())

		fmt.Println(common.Green("Record deleted successfully"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.ValidArgsFunction = serviceCompletion
}
