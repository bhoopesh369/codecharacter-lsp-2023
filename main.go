package main

import (
	"github.com/delta/codecharacter-lsp-2023/config"
	"github.com/delta/codecharacter-lsp-2023/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	server := echo.New()
	utils.InitLogger(server)
	server.Use(middleware.CORS())
	config.InitConfig()

	server.Logger.Fatal(server.Start(":" + config.ServerPort))
}