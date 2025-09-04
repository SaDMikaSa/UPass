package main

import (
	"fmt"

	"github.com/SaDMikaSa/UPass/config/configmanage"
	"github.com/SaDMikaSa/UPass/config/inputdata"
	"github.com/SaDMikaSa/UPass/internal/color"
)

func main() {
	color.Init()
	color.WelcomeBanner()

	cfg, err := configmanage.InitializeConfig()
	if err != nil {
		fmt.Errorf("failed to initialize config: %w", err)
	}

	// Показываем информацию о конфигурации
	inputdata.PrintConfigInfo(cfg)

	color.PrintInfo("Application is ready to use!")
}
