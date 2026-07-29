package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
)

func generatePasswordWithCharset(charset string, length int) ([]byte, error) {
	if len(charset) == 0 {
		return nil, common.ErrCharsetEmpty
	}
	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range password {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			common.ZeroBytes(password)
			return nil, fmt.Errorf("crypto/rand failed: %v", err)
		}
		password[i] = charset[n.Int64()]
	}
	return password, nil
}

// interactiveGenerate generates a password and allows the user to specify
// forbidden characters. It regenerates the password until the user accepts it.
func interactiveGenerate() ([]byte, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if !genNoSymbols {
		charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
	}

	for {
		password, err := generatePasswordWithCharset(charset, genLength)
		if err != nil {
			return nil, err
		}
		common.CyanPrintf("Generated password: %s\n", password)

		fmt.Print("Enter forbidden characters (e.g., O0l1, or \"O 0\" to forbid spaces), or Enter to keep: ")
		input, err := readLine()
		if err != nil || strings.TrimSpace(input) == "" {
			return password, nil
		}

		var forbiddenStr string
		if (strings.HasPrefix(input, `"`) && strings.HasSuffix(input, `"`)) ||
			(strings.HasPrefix(input, `'`) && strings.HasSuffix(input, `'`)) {
			unquoted, err := strconv.Unquote(input)
			if err == nil {
				forbiddenStr = unquoted
			} else {
				forbiddenStr = input[1 : len(input)-1]
			}
		} else {
			forbiddenStr = strings.ReplaceAll(input, " ", "")
		}

		if forbiddenStr != "" {
			var newCharset strings.Builder
			for _, ch := range charset {
				if !strings.ContainsRune(forbiddenStr, ch) {
					newCharset.WriteRune(ch)
				}
			}
			charset = newCharset.String()

			if len(charset) == 0 {
				common.RedPrintln("Error: All characters have been excluded! Resetting to default.")
				charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
				if !genNoSymbols {
					charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
				}
				continue
			}
		}

		hasForbidden := false
		for _, r := range string(password) {
			if strings.ContainsRune(forbiddenStr, r) {
				hasForbidden = true
				break
			}
		}

		if !hasForbidden {
			common.GreenPrintln("None of the forbidden characters were in the password. Keeping current.")
			return password, nil
		}

		common.YellowPrintln("Regenerating password with updated rules...")
		common.ZeroBytes(password)
	}
}

func generateRandomPassword() ([]byte, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if !genNoSymbols {
		charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
	}
	if genExclude != "" {
		var filtered strings.Builder
		for _, ch := range charset {
			if !strings.ContainsRune(genExclude, ch) {
				filtered.WriteRune(ch)
			}
		}
		charset = filtered.String()
	}

	result, err := generatePasswordWithCharset(charset, genLength)
	if err != nil {
		return nil, err
	}
	return result, nil
}
