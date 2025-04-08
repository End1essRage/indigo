package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &FileStorage{basePath: basePath}, nil
}

func (fs *FileStorage) getPath(entityType string, id string) string {
	return filepath.Join(fs.basePath, entityType, id+".json")
}

func (fs *FileStorage) Save(ctx context.Context, entityType string, id string, data interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	path := fs.getPath(entityType, id)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (fs *FileStorage) Load(ctx context.Context, entityType string, id string, result interface{}) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	path := fs.getPath(entityType, id)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(result)
}
