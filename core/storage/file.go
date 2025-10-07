package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory", err)
	}
	return &FileStorage{basePath: basePath}, nil
}

func (fs *FileStorage) getPath(docFolder, docPath string) string {
	return filepath.Join(fs.basePath, docFolder, docPath)
}

func (fs *FileStorage) Get(ctx context.Context, collection string, count int, query string) ([]Entity, error) {
	return nil, nil
}

func (fs *FileStorage) GetById(ctx context.Context, collection string, id string) (Entity, error) {
	// Проверка контекста
	if err := ctx.Err(); err != nil {
		return NewEntity(), fmt.Errorf("context error: %w", err)
	}

	errChan := make(chan error, 1)
	resultChan := make(chan Entity, 1)

	go func() {
		var result Entity

		if err := fs.load(collection, id, &result); err != nil {
			errChan <- fmt.Errorf("Ошбика загрзуки сущности", err)
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return NewEntity(), fmt.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errChan:
		return NewEntity(), err
	case data := <-resultChan:
		return data, nil
	}
}

func (fs *FileStorage) Create(ctx context.Context, collection string, entity Entity) (string, error) {
	// Проверка контекста
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context error: %w", err)
	}

	errChan := make(chan error, 1)
	resultChan := make(chan string, 1)

	go func() {
		id, err := fs.save(collection, entity)
		if err != nil {
			errChan <- fmt.Errorf("Ошбика загрзуки сущности", err)
			return
		}
		resultChan <- id
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errChan:
		return "", err
	case data := <-resultChan:
		return data, nil
	}
}

func (fs *FileStorage) UpdateById(ctx context.Context, collection string, id string, entity Entity) error {
	return nil
}

func (fs *FileStorage) Update(ctx context.Context, collection string, query string, entity Entity) (int, error) {
	return 0, nil
}

func (fs *FileStorage) DeleteById(ctx context.Context, collection string, id string) error {
	return nil
}

func (fs *FileStorage) Delete(ctx context.Context, collection string, query string) (int, error) {
	return 0, nil
}

func (fs *FileStorage) save(docFolder string, data Entity) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	id := uuid.New().ID()
	path := fs.getPath(docFolder, fmt.Sprint(id))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		return "", err
	}

	return fmt.Sprint(id), nil
}

func (fs *FileStorage) load(docFolder, docPath string, result *Entity) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	path := fs.getPath(docFolder, docPath)

	// Если файла нет - возвращаем nil без ошибки
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logrus.Debug("no file")
		return nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("нельзя передавать папку")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, _ := os.ReadFile(path)
	logrus.Debugf("data is %s", bytes)

	return json.NewDecoder(file).Decode(result)
}
