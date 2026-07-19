package cmd

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/health"
	"github.com/spf13/cobra"
)

var (
	skipHIBP bool
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run vault security health check",
	Long: `Check your passwords for weaknesses, duplicates, breaches, and reused logins.
			Breached password check uses Have I Been Pwned API with k-anonymity:
			only the first 5 characters of the SHA-1 hash are sent over the network.
			Use --no-hibp to skip this check.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		if err != nil {
			return err
		}
		defer common.ZeroBytes(password)

		fmt.Println(common.Cyan("======================"))
		fmt.Println(common.Cyan("Vault Health Report"))
		fmt.Println(common.Cyan("======================"))
		fmt.Println()

		records := vaultService.Records()
		if len(records) == 0 {
			fmt.Println(common.Yellow("Vault is empty. Nothing to check."))
			return nil
		}

		fmt.Println(common.Yellow("⚠️ WEAK PASSWORDS (strength score < %d out of 4)", common.MinStrengthScore))
		weakPasswords := health.CheckWeakPasswords(records, common.MinStrengthScore)
		if len(weakPasswords) == 0 {
			fmt.Println(common.Green("All passwords are strong!"))
		} else {
			for _, w := range weakPasswords {
				fmt.Printf("  %s — score: %d/4, crack time: %s\n",
					common.Red(w.Service),
					w.Score,
					w.CrackTime,
				)
			}
		}
		fmt.Println()

		fmt.Println(common.Yellow("DUPLICATE PASSWORDS"))
		duplicates := health.CheckDuplicatePasswords(records)
		if len(duplicates) == 0 {
			fmt.Println(common.Green("	No duplicate passwords!"))
		} else {
			for _, d := range duplicates {
				fmt.Printf("  %d services share the same password:\n", len(d.Services))
				for _, s := range d.Services {
					fmt.Printf("    - %s\n", common.Cyan(s))
				}
			}
		}
		fmt.Println()

		var breached []health.BreachedResult
		if skipHIBP {
			fmt.Println(common.Yellow("BREACHED PASSWORDS — skipped (--no-hibp)"))
		} else {
			fmt.Println(common.Red("BREACHED PASSWORDS (via Have I Been Pwned)"))
			fmt.Println("  Checking... (this may take a moment)")
			breached = health.CheckAllBreached(records)
			if len(breached) == 0 {
				fmt.Println(common.Green("	No breached passwords found!"))
			} else {
				for _, b := range breached {
					fmt.Printf("  %s — found in %s data breaches!\n",
						common.Red("%s", b.Service),
						common.Red("%d", b.Count),
					)
				}
			}
		}
		fmt.Println()

		fmt.Println(common.Yellow("REUSED LOGINS"))
		reused := health.CheckReusedLogins(records)
		if len(reused) == 0 {
			fmt.Println(common.Green("	All logins are unique!"))
		} else {
			for _, r := range reused {
				fmt.Printf("  %s used in %d services:\n", r.Login, len(r.Services))
				for _, s := range r.Services {
					fmt.Printf("    - %s\n", common.Cyan(s))
				}
			}
		}
		fmt.Println()

		fmt.Println(common.Cyan("SUMMARY"))
		fmt.Printf("  Total records: %d\n", len(records))
		fmt.Printf("  Weak passwords: %d\n", len(weakPasswords))
		fmt.Printf("  Duplicate groups: %d\n", len(duplicates))
		if !skipHIBP {
			fmt.Printf("  Breached passwords: %d\n", len(breached))
		}
		fmt.Printf("  Reused logins: %d\n", len(reused))

		return nil
	},
}

func init() {
	healthCmd.Flags().BoolVar(&skipHIBP, "no-hibp", false, "Skip breached passwords check (offline mode)")
	rootCmd.AddCommand(healthCmd)
}
