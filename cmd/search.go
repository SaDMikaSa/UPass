package cmd

import (
	"fmt"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/spf13/cobra"
)

var (
	searchLogin bool
	searchNote  bool
	searchTag   bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search services, logins, notes, and tags",
	Long: `Search for records matching the query.
			By default searches only in service names.
			Use --login, --note, and --tag flags to extend search.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		query := strings.ToLower(args[0])

		if !cmd.Flags().Changed("login") && !cmd.Flags().Changed("note") && !cmd.Flags().Changed("tag") {
			searchLogin = true
			searchNote = true
		}

		results := vaultService.SearchAll(query, searchLogin, searchNote, searchTag)

		if len(results) == 0 {
			fmt.Println(common.Yellow("No results matching %q", query))
			return nil
		}

		fmt.Println(common.Cyan("Results for %q:", query))
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
				common.Cyan(r.Service),
				r.Login,
				noteIndicator,
				matchedInfo,
			)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVarP(&searchLogin, "login", "l", false, "Search in login fields")
	searchCmd.Flags().BoolVarP(&searchNote, "note", "n", false, "Search in notes")
	searchCmd.Flags().BoolVarP(&searchTag, "tag", "t", false, "Search in service tags (part after ':')")
	rootCmd.AddCommand(searchCmd)
}
