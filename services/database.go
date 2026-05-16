package services

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService() (*DatabaseService, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(home, ".clipboard-wails")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(appDir, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initSchema(db); err != nil {
		return nil, err
	}

	return &DatabaseService{db: db}, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS clipboard_items (
			id        TEXT PRIMARY KEY,
			content   TEXT NOT NULL,
			type      TEXT NOT NULL,
			timestamp DATETIME NOT NULL
		);
	`)
	return err
}

// SaveItem persists a new clipboard item, deduplicating by content+type.
func (s *DatabaseService) SaveItem(item ClipboardItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	// Remove any existing entry with the same content+type so it floats to top.
	_, _ = s.db.Exec(
		`DELETE FROM clipboard_items WHERE content = ? AND type = ?`,
		item.Content, string(item.Type),
	)

	_, err := s.db.Exec(
		`INSERT INTO clipboard_items (id, content, type, timestamp) VALUES (?, ?, ?, ?)`,
		item.ID, item.Content, string(item.Type), item.Timestamp,
	)
	return err
}

// GetRecentItems returns the most recent limit items, newest first.
func (s *DatabaseService) GetRecentItems(limit int) ([]ClipboardItem, error) {
	rows, err := s.db.Query(
		`SELECT id, content, type, timestamp FROM clipboard_items ORDER BY timestamp DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ClipboardItem
	for rows.Next() {
		var item ClipboardItem
		var typeStr string
		var ts time.Time
		if err := rows.Scan(&item.ID, &item.Content, &typeStr, &ts); err != nil {
			continue
		}
		item.Type = ClipboardItemType(typeStr)
		item.Timestamp = ts
		items = append(items, item)
	}
	return items, nil
}

// DeleteItem removes a single item by ID.
func (s *DatabaseService) DeleteItem(id string) error {
	_, err := s.db.Exec(`DELETE FROM clipboard_items WHERE id = ?`, id)
	return err
}

// ClearAll removes every item from history.
func (s *DatabaseService) ClearAll() error {
	_, err := s.db.Exec(`DELETE FROM clipboard_items`)
	return err
}
