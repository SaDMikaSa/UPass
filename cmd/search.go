package cmd

import (
	"fmt"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search services, logins, notes, and tags",
	Long: `Search for records matching the query. By default searches only in service names. 
	Use --login, --note, and --tag flags to extend search.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		query := strings.ToLower(args[0])

		inLogin, _ := cmd.Flags().GetBool("login")
		inNote, _ := cmd.Flags().GetBool("note")
		inTag, _ := cmd.Flags().GetBool("tag")

		if !cmd.Flags().Changed("login") && !cmd.Flags().Changed("note") && !cmd.Flags().Changed("tag") {
			inLogin = true
			inNote = true
		}

		results := vaultService.SearchAll(query, inLogin, inNote, inTag)

		if len(results) == 0 {
			common.RedPrintf("No results matching %q\n", query)
			return nil
		}

		common.CyanPrintf("Results for %q:\n", query)
		for _, r := range results {
			noteIndicator := ""
			if vaultService.HasNote(r.Service) {
				noteIndicator = " *"
			}

			matchedInfo := ""
			if r.MatchedIn != "service" {
				matchedInfo = fmt.Sprintf(" (found in %s)", r.MatchedIn)
			}

			fmt.Printf("  %s | %s%s%s\n",
				r.Service,
				r.Login,
				noteIndicator,
				matchedInfo,
			)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolP("login", "l", false, "Search in login fields")
	searchCmd.Flags().BoolP("note", "n", false, "Search in service notes")
	searchCmd.Flags().BoolP("tag", "t", false, "Search in service tags")
	rootCmd.AddCommand(searchCmd)
}
