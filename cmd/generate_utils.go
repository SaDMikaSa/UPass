package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
)

func generatePasswordWithCharset(charset string, length int) []byte {
	if len(charset) == 0 {
		panic("generatePasswordWithCharset: charset is empty")
	}

	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range password {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			common.ZeroBytes(password)
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		password[i] = charset[n.Int64()]
	}
	return password
}

// interactiveGenerate generates a password and allows the user to specify
// forbidden characters. It regenerates the password until the user accepts it.
func interactiveGenerate() []byte {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if !genNoSymbols {
		charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
	}

	for {
		password := generatePasswordWithCharset(charset, genLength)
		fmt.Printf("Generated password: %s\n", string(password))

		fmt.Print("Enter forbidden characters (e.g., O0l1, or \"O 0\" to forbid spaces), or Enter to keep: ")
		input, err := readLine()
		if err != nil || strings.TrimSpace(input) == "" {
			return password
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
				fmt.Println(common.Red("Error: All characters have been excluded! Resetting to default."))
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
			fmt.Println(common.Green("None of the forbidden characters were in the password. Keeping current."))
			return password
		}

		fmt.Println(common.Yellow("Regenerating password with updated rules..."))
		common.ZeroBytes(password)
	}
}

func generateRandomPassword() []byte {
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
	return generatePasswordWithCharset(charset, genLength)
}
