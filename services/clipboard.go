package services

import (
	"context"
	"fmt"
	"time"

	"golang.design/x/clipboard"
)

// ClipboardService monitors the system clipboard and persists new items.
type ClipboardService struct {
	db        *DatabaseService
	storage   *StorageService
	OnNewItem func(ClipboardItem)
}

func NewClipboardService(db *DatabaseService, storage *StorageService) *ClipboardService {
	return &ClipboardService{db: db, storage: storage}
}

// Start initialises the clipboard library and launches background watchers.
func (s *ClipboardService) Start() error {
	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("clipboard init: %w", err)
	}
	go s.watchText()
	go s.watchImage()
	return nil
}

func (s *ClipboardService) watchText() {
	ch := clipboard.Watch(context.Background(), clipboard.FmtText)
	for data := range ch {
		text := string(data)
		if text == "" {
			continue
		}
		s.process(ClipboardItem{
			Content:   text,
			Type:      TypeClipboardText,
			Timestamp: time.Now(),
		})
	}
}

func (s *ClipboardService) watchImage() {
	ch := clipboard.Watch(context.Background(), clipboard.FmtImage)
	for data := range ch {
		if len(data) == 0 {
			continue
		}
		path, err := s.storage.SaveImage(data)
		if err != nil {
			fmt.Printf("clipboard: save image: %v\n", err)
			continue
		}
		s.process(ClipboardItem{
			Content:   path,
			Type:      TypeClipboardImage,
			Timestamp: time.Now(),
		})
	}
}

func (s *ClipboardService) process(item ClipboardItem) {
	if err := s.db.SaveItem(item); err != nil {
		fmt.Printf("clipboard: save item: %v\n", err)
		return
	}
	if s.OnNewItem != nil {
		s.OnNewItem(item)
	}
}

// Write puts a clipboard item back onto the system clipboard.
func (s *ClipboardService) Write(item ClipboardItem) {
	switch item.Type {
	case TypeClipboardText:
		clipboard.Write(clipboard.FmtText, []byte(item.Content))
	// Image write-back omitted: would require reading the stored file.
	}
}
