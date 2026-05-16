package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Clipboard Manager",
		Width:  780,
		Height: 560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 24, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose: func(ctx context.Context) bool {
			// Hide instead of quit unless the user chose Quit explicitly.
			if app.shouldQuit {
				return false
			}
			runtime.WindowHide(ctx)
			return true
		},
		StartHidden:           true,
		HideWindowOnClose:     false,
		Frameless:             false,
		EnableDefaultContextMenu: false,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
