package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &FileStorage{basePath: basePath}, nil
}

func (fs *FileStorage) getPath(docFolder, docPath string) string {
	return filepath.Join(fs.basePath, docFolder, docPath)
}

func (fs *FileStorage) Get(ctx context.Context, collection string, count int, query string) ([]Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
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
			errChan <- fmt.Errorf("ошибка загрузки сущности: %w", err)
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
		id, err := fs.save(0, collection, entity)
		if err != nil {
			errChan <- fmt.Errorf("ошибка сохранения сущности: %w", err)
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
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	var result Entity
	err := fs.load(collection, id, &result)
	if err != nil {
		return fmt.Errorf("ошибка поиска по айди: %w", err)
	}

	for k, v := range entity {
		result[k] = v
	}

	ids, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	_, err = fs.save(uint32(ids), collection, result)

	return nil
}

func (fs *FileStorage) Update(ctx context.Context, collection string, query string, entity Entity) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}
	return 0, nil
}

func (fs *FileStorage) DeleteById(ctx context.Context, collection string, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	return nil
}

func (fs *FileStorage) Delete(ctx context.Context, collection string, query string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}
	return 0, nil
}

// save с атомарной записью (БЕЗ МЬЮТЕКСА!)
func (fs *FileStorage) save(id uint32, docFolder string, data Entity) (string, error) {
	// Генерируем ID если нужно
	if id == 0 {
		id = uuid.New().ID()
	}

	path := fs.getPath(docFolder, fmt.Sprint(id))

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	// АТОМАРНАЯ ЗАПИСЬ: пишем во временный файл
	tempPath := path + ".tmp." + fmt.Sprint(time.Now().UnixNano())
	file, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	encErr := enc.Encode(data)
	file.Close() // Закрываем перед rename!

	if encErr != nil {
		os.Remove(tempPath)
		return "", encErr
	}

	// Атомарное переименование (гарантируется ОС)
	// Это защищает от чтения частично записанного файла
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	return fmt.Sprint(id), nil
}

// load читает файл (БЕЗ МЬЮТЕКСА!)
// Каждый os.Open создает отдельный file descriptor
func (fs *FileStorage) load(docFolder, docPath string, result *Entity) error {
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

	return json.NewDecoder(file).Decode(result)
}
