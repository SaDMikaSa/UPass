package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SaDMikaSa/UPass/internal/color"
)

// Constants for menu operations and application behavior.
const (
	createOp    = "1. Create record"
	getOp       = "2. Get record"
	listOp      = "3. Add record"
	changeOp    = "4. Change record"
	deleteOp    = "5. Delete record"
	upassOp     = "6. Change {{UPass}}"
	exitOp      = "0. Exit"
	forkMessage = "Please select an option:"
	waitTime    = 2 * time.Second
	maxItem     = 6 // the current maximum value of the option.
)

// MenuItem represents a menu item with its associated handler function.
type MenuItem struct {
	OptionNumber int
	Text         string
	Handler      func() error
}

// menuItems is the list of available menu options with their handlers.
var menuItems = []MenuItem{
	{1, createOp, createRecord},
	{2, getOp, getRecord},
	{3, listOp, addRecord},
	{4, changeOp, changeRecord},
	{5, deleteOp, deleteRecord},
	{6, upassOp, changeUPass},
	{0, exitOp, exit},
}

// RunApp starts and runs the application's main loop.
// It handles menu display, user input processing, and workflow control.
func RunApp() error {
	if err := InitStorage(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageManager.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		printMenu()

		input, err := readUserInput(reader)
		if err != nil {
			color.PrintRejected("Invalid input. Enter a number.")
			time.Sleep(waitTime)
			continue
		}

		if err := handleUserChoice(input); err != nil {
			color.PrintRejected("Invalid choice.Try again.")
			time.Sleep(waitTime)
			continue
		}

		repeated, err := isRepeated(reader)
		if err != nil {
			color.PrintRejected("Invalid input. Enter [yes] or [no].")
			time.Sleep(waitTime)
		}

		if repeated {
			continue
		} else {
			color.PrintInfo("Application closed.")
			break
		}
	}
	return nil
}

// printMenu displays the application menu using colored output.
// Shows all available menu options to the user.
func printMenu() {

	color.PrintInfo("\n" + forkMessage)
	for _, item := range menuItems {
		color.PrintInfo(item.Text)
	}
}

// readUserInput reads and validates user menu choice.
// Returns parsed integer choice or error if input is invalid.
func readUserInput(reader *bufio.Reader) (int, error) {

	color.PrintPrompts("your choice")

	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	if choice < 0 || choice > maxItem {
		return 0, fmt.Errorf("enter a number between 0 and %d", maxItem)
	}

	return choice, nil
}

// handleUserChoice executes the corresponding menu handler based on user selection.
// Returns error if selected option doesn't exist.
func handleUserChoice(choice int) error {

	menuMap := make(map[int]func() error)
	for _, item := range menuItems {
		menuMap[item.OptionNumber] = item.Handler
	}

	if handler, exists := menuMap[choice]; exists {
		handler()
		return nil
	}
	return fmt.Errorf("invalid choice: %d", choice)
}

// isRepeated checks if the user wants to continue using the application.
func isRepeated(reader *bufio.Reader) (bool, error) {

	color.PrintInfo("Something else?")
	color.PrintPrompts("[yes] or [no]")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	switch strings.TrimSpace(strings.ToLower(input)) {
	case "y", "yes", "нуы", "н", "д":
		return true, nil
	default:
		return false, nil
	}
}
