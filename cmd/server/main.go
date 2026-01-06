package main

import (
	"go-wordpress/internal/app"
	"go-wordpress/internal/config"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token in the format: Bearer <token>
func main() {
	config.LoadEnv()

	app.NewApp().Run()
}
