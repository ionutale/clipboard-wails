package main

import (
	"clipboard-wails/services"
	"clipboard-wails/tray"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct.
type App struct {
	ctx        context.Context
	db         *services.DatabaseService
	storage    *services.StorageService
	clipboard  *services.ClipboardService
	shouldQuit bool

	mu          sync.RWMutex
	recentItems []services.ClipboardItem // last 10 for tray menu
}

// NewApp creates and wires up the application.
func NewApp() *App {
	db, err := services.NewDatabaseService()
	if err != nil {
		fmt.Printf("db init: %v\n", err)
	}
	storage, err := services.NewStorageService()
	if err != nil {
		fmt.Printf("storage init: %v\n", err)
	}
	clip := services.NewClipboardService(db, storage)
	return &App{db: db, storage: storage, clipboard: clip}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Pre-load recent items for the tray menu.
	if items, err := a.db.GetRecentItems(10); err == nil {
		a.mu.Lock()
		a.recentItems = items
		a.mu.Unlock()
		tray.UpdateMenu(a.buildMenuLabels(items))
	}

	// Start the tray icon with callbacks.
	tray.SetupTray(
		ctx,
		func(idx int) { a.trayItemClicked(idx) },
		func() { a.ClearHistory() },
		func() { a.Quit() },
	)

	// Wire up new-item callback.
	a.clipboard.OnNewItem = func(item services.ClipboardItem) {
		// Refresh in-memory recent list.
		a.mu.Lock()
		a.recentItems = prepend(item, a.recentItems, 10)
		labels := a.buildMenuLabels(a.recentItems)
		a.mu.Unlock()

		tray.UpdateMenu(labels)
		runtime.EventsEmit(a.ctx, "clipboard-new-item", item)
	}

	if err := a.clipboard.Start(); err != nil {
		fmt.Printf("clipboard start: %v\n", err)
	}
}

// ---- public API (bound to JS) ----

// GetHistory returns the 200 most recent clipboard items.
func (a *App) GetHistory() []services.ClipboardItem {
	items, err := a.db.GetRecentItems(200)
	if err != nil {
		return []services.ClipboardItem{}
	}
	return items
}

// CopyToClipboard writes an item back onto the system clipboard.
func (a *App) CopyToClipboard(item services.ClipboardItem) {
	a.clipboard.Write(item)
}

// DeleteItem removes a single history entry.
func (a *App) DeleteItem(id string) error {
	return a.db.DeleteItem(id)
}

// ClearHistory deletes all clipboard history.
func (a *App) ClearHistory() error {
	err := a.db.ClearAll()
	if err == nil {
		a.mu.Lock()
		a.recentItems = nil
		a.mu.Unlock()
		tray.UpdateMenu(nil)
		runtime.EventsEmit(a.ctx, "clipboard-cleared")
	}
	return err
}

// GetImageData returns a base64-encoded PNG data-URL for an image item.
// The path must be inside ~/.clipboard-wails/images/ to prevent path traversal.
func (a *App) GetImageData(path string) (string, error) {
	allowedDir := a.storage.DataDir()
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, allowedDir) {
		return "", fmt.Errorf("forbidden path")
	}
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
}

// Quit exits the application.
func (a *App) Quit() {
	a.shouldQuit = true
	runtime.Quit(a.ctx)
}

// ---- helpers ----

func (a *App) trayItemClicked(idx int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if idx < 0 || idx >= len(a.recentItems) {
		return
	}
	a.clipboard.Write(a.recentItems[idx])
	runtime.WindowShow(a.ctx)
}

func (a *App) buildMenuLabels(items []services.ClipboardItem) []string {
	labels := make([]string, len(items))
	for i, item := range items {
		switch item.Type {
		case services.TypeClipboardImage:
			labels[i] = "🖼️  [Image]"
		default:
			s := item.Content
			if len(s) > 60 {
				s = s[:57] + "..."
			}
			labels[i] = "📝  " + s
		}
	}
	return labels
}

// prepend inserts item at the front of the slice, deduplicating by content,
// and caps the length at max.
func prepend(item services.ClipboardItem, items []services.ClipboardItem, max int) []services.ClipboardItem {
	result := make([]services.ClipboardItem, 0, max)
	result = append(result, item)
	for _, existing := range items {
		if existing.Content == item.Content {
			continue
		}
		result = append(result, existing)
		if len(result) >= max {
			break
		}
	}
	return result
}

