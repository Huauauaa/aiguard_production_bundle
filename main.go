package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "AI代码监视",
		Width:            1360,
		Height:           920,
		MinWidth:         1120,
		MinHeight:        760,
		DisableResize:    false,
		Frameless:        false,
		BackgroundColour: &options.RGBA{R: 17, G: 24, B: 39, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
	})
	if err != nil {
		panic(err)
	}
}
