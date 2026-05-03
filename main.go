package main

import (
	"embed"
	"log"

	"pacto/internal/audio"
	"pacto/internal/state"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	appState := state.New()
	if err := audio.Init(); err != nil {
		log.Fatal(err)
	}
	app := NewApp(appState)

	err := wails.Run(&options.App{
		Title:         "PACTO",
		Width:         420,
		Height:        300,
		MinWidth:      420,
		MinHeight:     300,
		DisableResize: false,
		AlwaysOnTop:   true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
