package tray

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SetupTray initialises the system tray.
// onItemClick is called with the 0-based index of the clicked history item.
// onClear is called when "Clear History" is chosen.
// onQuit is called when "Quit" is chosen.
func SetupTray(ctx context.Context, onItemClick func(int), onClear func(), onQuit func()) {
	setupTrayImpl(
		func() { runtime.WindowShow(ctx) },
		onItemClick,
		onClear,
		onQuit,
	)
}

// UpdateMenu refreshes the dropdown list of recent clipboard item labels.
// items should be short display strings (truncated text or "[Image]").
func UpdateMenu(items []string) {
	updateMenuImpl(items)
}
