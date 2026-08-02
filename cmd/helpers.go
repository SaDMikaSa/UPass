package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/SaDMikaSa/UPass/pkg/tyuiop"
	"golang.org/x/term"
)

// unlock prompts the user for the master password and unlocks the vault
// service. On success it returns the entered password (caller should zero it
// when finished) and the service will be in unlocked state.
func unlock() ([]byte, error) {
	password, err := inputPass("Enter master password: ")
	if err != nil {
		return nil, fmt.Errorf("unlock: %w", err)
	}

	if err := vaultService.Unlock(password); err != nil {
		common.ZeroBytes(password)
		return nil, fmt.Errorf("unlock: %w", err)
	}
	return password, nil
}

// inputPass reads a password from the terminal without echoing. The prompt
// is printed to stdout. Returns ErrNotEmpty when the user enters an empty
// password.
func inputPass(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("error reading password: %w", err)
	}

	if len(pass) == 0 {
		return nil, common.ErrNotEmpty
	}

	fmt.Println()

	if err := common.LockMemory(pass); err != nil {
		common.ZeroBytes(pass)
		return nil, fmt.Errorf("could not lock memory for password (mlock failed: %w). Aborting to prevent swap leakage", err)
	}

	return pass, nil
}

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", fmt.Errorf("EOF")
		}
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// readConfirmation reads a yes/no confirmation from stdin and returns true
// if the user answered 'y' or 'Y'. Any error is treated as 'no'.
func readConfirmation() bool {
	ans, err := readLine()
	if err != nil {
		return false
	}
	return ans == "y" || ans == "Y"
}

func saveServicesCache(services []string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	cachePath := filepath.Join(home, store.DefaultVaultDir, cacheFileName)
	data := strings.Join(services, "\n")

	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create cache directory %s: %v\n", cachePath, err)
		return
	}

	if err := os.WriteFile(cachePath, []byte(data), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save completion cache to %s: %v\n", cachePath, err)
	}
}

// promptNewMasterPassword reads and confirms a new master password.
// Returns nil if user cancels on weak password warning.
func newMasterPassword() ([]byte, error) {
	pass, err := inputPass("Enter master password: ")
	if err != nil {
		return nil, err
	}

	strength := tyuiop.AnalyzeLocalStrength(pass)
	if strength.Score < common.MinStrengthScore {
		common.YellowPrintf("⚠️  Weak password (score: %d/4, crack time: %s)\n", strength.Score, strength.CrackTime)
		fmt.Print("Continue anyway? (y/n): ")

		if !readConfirmation() {
			common.ZeroBytes(pass)
			return nil, common.ErrCanceled
		}
	}

	return pass, nil
}
