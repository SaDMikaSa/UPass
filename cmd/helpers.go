package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
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

	return pass, nil
}

// readLine reads a single trimmed line from stdin. It returns an error on
// EOF or when the input scanner produces an error.
func readLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		return "", fmt.Errorf("EOF")
	}
	return strings.TrimSpace(scanner.Text()), nil
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

// saveServicesCache persists a small cache of service names into the
// vault directory to speed up shell completion. This cache contains no
// secrets and is stored with restrictive permissions.
func saveServicesCache(services []string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	cachePath := filepath.Join(home, store.DefaultVaultDir, cacheFileName)
	data := strings.Join(services, "\n")

	os.MkdirAll(filepath.Dir(cachePath), 0700)
	os.WriteFile(cachePath, []byte(data), 0600)
}
