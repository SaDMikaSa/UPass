package app

import (
	"bufio"
	"os"
	"strings"

	"github.com/SaDMikaSa/UPass/config/inputdata"
	"github.com/SaDMikaSa/UPass/internal/color"
)

var reader = bufio.NewReader(os.Stdin)

func createRecord() {
	color.PrintInfo("CreateData...")
	for {
		color.PrintPrompts("service name")
		url, err := reader.ReadString('\n')
		url = strings.TrimSpace(url)
		if err != nil || url == "" {
			color.PrintRejected("Invalid input")
			continue
		}

		color.PrintPrompts("username/email")
		login, err := reader.ReadString('\n')
		login = strings.TrimSpace(login)
		if err != nil || login == "" {
			color.PrintRejected("Invalid input")
			continue
		}

		color.PrintPrompts("account password")
		pass, err := reader.ReadString('\n')
		pass = strings.TrimSpace(pass)
		if err != nil || pass == "" {
			color.PrintRejected("Invalid input")
			continue
		}

		color.PrintSuccess("Record successfully created")
		color.PrintInfo(url + " " + login + " " + pass)
		break
	}
}

func addRecord() {
	color.PrintInfo("AddData...")
}

func getRecord() {
	color.PrintInfo("GetData...")
	color.PrintPrompts("service name to search")
	service, err := reader.ReadString('\n')
	if err != nil || service == "" {
		color.PrintRejected("Invalid input")
	}
	service = strings.TrimSpace(service)

}

func changeRecord() {
	color.PrintInfo("ChangeData...")
}

func deleteRecord() {
	color.PrintInfo("DeleteData...")
}

func exit() {
	color.PrintInfo("App is closed")
	os.Exit(1)
}

func changeUPass() {
	color.PrintPrompts("{{UPass}}")
	inputdata.EnterPassword()
}
