package app

import (
	"os"

	"github.com/SaDMikaSa/UPass/internal/color"
)

func createData() {
	color.PrintInfo("CreateData...")
}

func addData() {
	color.PrintInfo("AddData...")
}

func getData() {
	color.PrintInfo("GetData...")
}

func changeData() {
	color.PrintInfo("ChangeData...")
}

func deleteData() {
	color.PrintInfo("DeleteData...")
}

func exit() {
	color.PrintInfo("App is closed")
	os.Exit(1)
}
