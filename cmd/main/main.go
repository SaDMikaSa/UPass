package main

import (
	"log"
	"os"

	"github.com/SaDMikaSa/UPass/cmd/app"
	"github.com/SaDMikaSa/UPass/config/configmanage"
	"github.com/SaDMikaSa/UPass/internal/auth"
	"github.com/SaDMikaSa/UPass/internal/color"
)

func main() {
	color.Init()
	color.WelcomeBanner()

	_, err := configmanage.InitializeConfig()
	if err != nil {
		log.Printf("Failed to initialize config: %v", err)
		os.Exit(1)
	}
	IsAutarization, err := auth.IsAuthenticated()
	if err != nil {
		log.Printf("Failed to authenticate: %v", err)
	}
	if IsAutarization {
		color.PrintSuccess("Authenticated successfully ")
		app.RunApp()
	} else {
		color.PrintRejected("Authentication failed. Please try again.")
	}

}
