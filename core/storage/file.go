package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
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

func (fs *FileStorage) getPath(docFolder, docPath string) string {
	return filepath.Join(fs.basePath, docFolder, docPath)
}

func (fs *FileStorage) Save(docFolder, docPath string, data interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	path := fs.getPath(docFolder, docPath)
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

func (fs *FileStorage) Load(docFolder, docPath string, result interface{}) error {
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

func (fs *FileStorage) LoadArray(docFolder, docPath string) ([]interface{}, error) {
	result := make([]interface{}, 0)

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	path := fs.getPath(docFolder, docPath)

	// Если файла нет - возвращаем nil без ошибки
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logrus.Debug("no file")
		return result, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return result, err
	}

	if !fi.IsDir() {
		return result, fmt.Errorf("ожидалась папка")
	}

	ents, err := os.ReadDir(path)
	if err != nil {
		return result, err
	}

	for _, ent := range ents {
		file, err := os.Open(filepath.Join(path, ent.Name()))
		if err != nil {
			return result, err
		}
		defer file.Close()

		var res interface{}

		json.NewDecoder(file).Decode(&res)

		result = append(result, res)
	}

	return result, nil
}
