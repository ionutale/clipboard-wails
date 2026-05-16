package services

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type StorageService struct {
	dir string
}

func NewStorageService() (*StorageService, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".clipboard-wails", "images")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &StorageService{dir: dir}, nil
}

// SaveImage writes raw PNG bytes to disk and returns the absolute path.
func (s *StorageService) SaveImage(data []byte) (string, error) {
	path := filepath.Join(s.dir, uuid.New().String()+".png")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return path, nil
}

// DataDir exposes the storage directory so the app can validate paths.
func (s *StorageService) DataDir() string {
	return s.dir
}
