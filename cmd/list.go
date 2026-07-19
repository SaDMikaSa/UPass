package cmd

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		services := vaultService.ListServices()
		if services == nil {
			return common.ErrLocked
		}

		if len(services) == 0 {
			fmt.Println(common.Yellow("No records found"))
			return nil
		}

		fmt.Println(common.Cyan("Services:"))
		for i, s := range services {
			fmt.Printf("  %d. %s\n", i+1, common.Cyan(s))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
