// Command server runs the stateless product extraction service.
package main

import (
	"go-wordpress/internal/app"
	"go-wordpress/internal/config"
)

func main() {
	config.LoadEnv()
	app.NewApp().Run()
}
