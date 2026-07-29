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
	Long: `Check your passwords for weaknesses, duplicates, breaches, and reused logins. Breached password check uses Have I Been Pwned API with k-anonymity: 
	only the first 5 characters of the SHA-1 hash are sent over the network. 
	Use --no-hibp to skip this check.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := unlock()
		defer common.ZeroBytes(password)
		if err != nil {
			return err
		}

		common.CyanPrintln("Vault Health Report")
		fmt.Println()

		records := vaultService.Records()
		if len(records) == 0 {
			common.YellowPrintln("Vault is empty. Nothing to check.")
			return nil
		}

		common.YellowPrintf("WEAK PASSWORDS (strength score < %d out of 4)\n", common.MinStrengthScore)
		weakPasswords := health.CheckWeakPasswords(records, common.MinStrengthScore)
		if len(weakPasswords) == 0 {
			common.GreenPrintln("All passwords are strong!")
		} else {
			for _, w := range weakPasswords {
				common.RedPrintf("  %s — score: %d/4, crack time: %s\n",
					w.Service,
					w.Score,
					w.CrackTime,
				)
			}
		}
		fmt.Println()

		common.YellowPrintln("DUPLICATE PASSWORDS: ")
		duplicates := health.CheckDuplicatePasswords(records)
		if len(duplicates) == 0 {
			common.GreenPrintln(" No duplicate passwords!")
		} else {
			for _, d := range duplicates {
				fmt.Printf(" %d services share the same password:\n", len(d.Services))
				for _, s := range d.Services {
					common.CyanPrintf("    - %s\n", s)
				}
			}
		}
		fmt.Println()

		var breached []health.BreachedResult
		if skipHIBP {
			common.YellowPrintln("BREACHED PASSWORDS — skipped (--no-hibp)")
		} else {
			common.RedPrintln("BREACHED PASSWORDS (via Have I Been Pwned)")
			fmt.Println("Checking... (this may take a moment)")
			breached, err = health.CheckAllBreached(records)
			if err != nil {
				common.YellowPrintf("⚠️ Warning: Could not check breached passwords (%v).", err)
				common.YellowPrintln("  Your passwords were NOT sent to the server. Check your internet connection.")
			} else if len(breached) == 0 {
				common.GreenPrintln(" No breached passwords found!")
			} else {
				for _, b := range breached {
					common.RedPrintf("  %s — found in %d data breaches!\n",
						b.Service,
						b.Count,
					)
				}
			}
		}
		fmt.Println()

		common.YellowPrintln("REUSED LOGINS:")
		reused := health.CheckReusedLogins(records)
		if len(reused) == 0 {
			common.GreenPrintln("  All logins are unique!")
		} else {
			for _, r := range reused {
				fmt.Printf("  %s used in %d services:\n", r.Login, len(r.Services))
				for _, s := range r.Services {
					common.CyanPrintf("    - %s\n", s)
				}
			}
		}
		fmt.Println()

		common.CyanPrintln("SUMMARY")
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
