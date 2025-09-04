package inputdata

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/SaDMikaSa/UPass/config/models"
	"github.com/SaDMikaSa/UPass/internal/color"
	"golang.org/x/term"
)

const (
	maxAttempts     = 3
	lengthSeparator = 2
	minLength       = 12
	maxLength       = 30
	waitTime        = 30 * time.Second // 1 * time.Minute
)

var attemptsKey = map[int]string{
	2: "Attempts remaining: 2",
	1: "Attempts remaining: 1",
	0: "",
}

// checkAttempts for checking the number of attempts
func checkAttempts(attempts int, prompt func()) {
	if attempts > maxAttempts {
		color.PrintRejected("Too many wrong attempts, wait 1 minute and try again")
		time.Sleep(waitTime)
		attempts = 1
	}

	prompt()
}

// EnterPassword prompts the user to enter and confirm a password, enforcing complexity rules and retry limits.
func EnterPassword() (string, error) {
	attempts := 1

	for {
		checkAttempts(attempts, func() {
			color.PrintPrompts("{{UPass}}")
		})

		p, err := term.ReadPassword(0)
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}

		pass := strings.TrimSpace(string(p))
		if !isValidPassword(pass) {
			color.PrintInfo(attemptsKey[maxAttempts-attempts])
			attempts++
			continue
		}

		color.PrintInfo("")
		color.PrintPrompts("{{UPass}} again")

		p, err = term.ReadPassword(0)
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}

		if pass != strings.TrimSpace(string(p)) {
			color.PrintRejected("{{UPass}} doesn't match.Try again")
			color.PrintInfo(attemptsKey[maxAttempts-attempts])
			attempts++
			continue
		} else {
			color.PrintInfo("")

			return pass, nil
		}
	}
}

// isValidPassword validates password meets security requirements.
func isValidPassword(password string) bool {
	var (
		hasUpper, hasLower, hasDigit, hasSpecial = false, false, false, false
		mustContain                              = "{{UPass}} must contain at least one "
	)

	if len(password) < minLength || len(password) > maxLength {
		color.PrintRejected("{{UPass}} length must be greater than 12 and less than 30.Try again")
		return false
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		color.PrintRejected(mustContain + "one uppercase letter")
		return false
	}
	if !hasLower {
		color.PrintRejected(mustContain + "lowercase letter")
		return false
	}
	if !hasDigit {
		color.PrintRejected(mustContain + "digit")
		return false
	}
	if !hasSpecial {
		color.PrintRejected(mustContain + "special character")
		return false
	}

	return true
}

// EnterWord prompts the user for a secret word with a numeric suffix.
func EnterWord() (string, int, error) {
	reader := bufio.NewReader(os.Stdin)
	attempts := 1

	color.PrintInfo("There should be a number from 9 to 24 at the end of the secret word.")
	for {
		checkAttempts(attempts, func() {
			color.PrintPrompts("secret word")
		})

		color.PrintPrompts("secret word")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", 0, fmt.Errorf("failed to read word: %w", err)
		}
		word := strings.TrimSpace(input)

		if !isValidWord(strings.TrimSpace(word)) {
			color.PrintInfo(attemptsKey[maxAttempts-attempts])
			attempts++
			continue
		}
		wordPart := word[:len(word)-lengthSeparator]
		numPart := word[len(word)-lengthSeparator:]
		if !isNumeric(numPart) {
			color.PrintRejected("Last two characters must be digits")
			continue
		}
		number, err := strconv.Atoi(numPart)
		if err != nil || number < 9 || number > 24 {
			color.PrintRejected("Last two characters of your word must be integer from 9 to 24.Try again")
			continue
		}
		return wordPart, number, nil
	}
}

// isValidWord verifies the correctness of the entered secret word.
func isValidWord(word string) bool {
	if len(word) < minLength-lengthSeparator || len(word) > maxLength-lengthSeparator {
		color.PrintRejected("{{UPass}} length must be greater than 12 and less than 30.Try again")
		return false
	}
	return true
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func PrintConfigInfo(config *models.Config) {
	fmt.Println("\n=== Current Configuration ===")
	fmt.Printf("First run: %v\n", config.IsFirstRun)
	fmt.Printf("Power level: %d\n", config.Power)
	fmt.Println(config.Password)
	fmt.Println(config.Word)
	fmt.Println("============================")
}
