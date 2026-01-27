package main

import (
	"embed"
	_ "embed"
	"log"

	"changeme/internal/bootstrap"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/sysicon.png
var icon []byte

func init() {
	// application.RegisterEvent[string]("time")
}

func main() {
	app, err := bootstrap.NewApp(bootstrap.Options{
		Assets: assets,
		Icon:   icon,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
