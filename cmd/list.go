package cmd

import (
	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		services := vaultService.ListServices()
		if services == nil {
			return common.ErrLocked
		}

		if len(services) == 0 {
			common.RedPrintln("No records found")
			return nil
		}

		common.CyanPrintln("Services:")
		for i, s := range services {
			common.CyanPrintf("  %d. %s\n", i+1, s)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
