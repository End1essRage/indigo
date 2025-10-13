package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func (fs *FileStorage) Get(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	collectionPath := filepath.Join(fs.basePath, collection)

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("коллекция не существует: %w", err)
	}

	files, err := fs.listCollectionFiles(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return nil, NewNotFoundError(query.ToString())
	}

	results := make([]Entity, 0, count)

	for _, fileName := range files {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		entity, err := fs.loadAndFilter(ctx, collection, fileName, query)
		if err != nil {
			logrus.Errorf("ошибка фильтрации сущности %s", err.Error())
			continue
		}

		if entity != nil {
			results = append(results, *entity)
			if count > 0 && len(results) >= count {
				break
			}
		}
	}

	return results, nil
}

func (fs *FileStorage) GetOne(ctx context.Context, collection string, query QueryNode) (Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	collectionPath := filepath.Join(fs.basePath, collection)

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("коллекция не существует: %w", err)
	}

	files, err := fs.listCollectionFiles(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return nil, NewNotFoundError("в коллекции нет ни одной сущности")
	}

	for _, fileName := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		entity, err := fs.loadAndFilter(ctx, collection, fileName, query)
		if err != nil {
			logrus.Errorf("ошибка фильтрации сущности %s", err.Error())
			continue
		}

		if entity != nil {
			return *entity, nil
		}
	}

	return nil, NewNotFoundError(query.ToString())
}

func (fs *FileStorage) GetIds(ctx context.Context, collection string, count int, query QueryNode) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	result := make([]string, 0)

	items, err := fs.Get(ctx, collection, count, query)
	if err != nil {
		return nil, err
	}

	for _, e := range items {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		for k, v := range e {
			if k == "_id" {
				result = append(result, v.(string))
				continue
			}
		}
	}

	return result, nil
}

func (fs *FileStorage) GetById(ctx context.Context, collection string, id string) (Entity, error) {
	// Проверка контекста
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	errChan := make(chan error, 1)
	resultChan := make(chan Entity, 1)

	go func() {
		var result Entity
		if err := fs.load(ctx, collection, id, &result); err != nil {
			errChan <- NewNotFoundError("no found")
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errChan:
		return nil, err
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
		id, err := fs.save(ctx, "", collection, entity)
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
		logrus.Errorf("err creation with id %s", err.Error())
		return "", err
	case data := <-resultChan:
		logrus.Debugf("created with id %s", data)
		return data, nil
	}
}

func (fs *FileStorage) UpdateById(ctx context.Context, collection string, id string, entity Entity) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	errChan := make(chan error, 1)
	doneChan := make(chan struct{}, 1)

	go func() {
		var result Entity
		err := fs.load(ctx, collection, id, &result)
		if err != nil {
			errChan <- err
			return
		}

		for k, v := range entity {
			result[k] = v
		}

		_, err = fs.save(ctx, id, collection, result)
		if err != nil {
			errChan <- fmt.Errorf("ошибка сохранения: %w", err)
			return
		}

		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}

func (fs *FileStorage) Update(ctx context.Context, collection string, query QueryNode, entity Entity) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}

	ids, err := fs.GetIds(ctx, collection, 0, query)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			return 0, err
		} else {
			logrus.Error("непредвиденная ошибка %w", err)
			return 0, fmt.Errorf("ошибка получения списка сущностей: %w", err)
		}
	}

	counter := 0

	for _, id := range ids {
		select {
		case <-ctx.Done():
			return counter, ctx.Err()
		default:
		}

		if err := fs.UpdateById(ctx, collection, id, entity); err != nil {
			logrus.Errorf("ошибка обновления сущности id:%s", id)
			continue
		}

		counter++
	}

	return counter, nil
}

func (fs *FileStorage) DeleteById(ctx context.Context, collection string, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	err := fs.delete(ctx, collection, id)
	if err != nil {
		logrus.Errorf("[STORAGE] error %s", err.Error())
		return err
	}

	logrus.Debugf("[STORAGE] удалено %s", id)

	return err
}

func (fs *FileStorage) Delete(ctx context.Context, collection string, query QueryNode) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}

	ids, err := fs.GetIds(ctx, collection, 0, query)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			return 0, err
		} else {
			logrus.Error("непредвиденная ошибка %w", err)
			return 0, fmt.Errorf("ошибка получения списка сущностей: %w", err)
		}
	}

	counter := 0

	for _, id := range ids {
		select {
		case <-ctx.Done():
			return counter, ctx.Err()
		default:
		}

		if err := fs.DeleteById(ctx, collection, id); err != nil {
			logrus.Errorf("ошибка удаления сущности id:%s", id)
			continue
		}

		counter++
	}

	return counter, nil
}

// loadAndFilter загружает entity и применяет фильтр
func (fs *FileStorage) loadAndFilter(ctx context.Context, collection, fileName string, query QueryNode) (*Entity, error) {
	var entity Entity
	if err := fs.load(ctx, collection, fileName, &entity); err != nil {
		return nil, err
	}

	// Если query nil - возвращаем все
	if query == nil {
		return &entity, nil
	}

	// Применяем фильтр
	match, err := query.Evaluate(entity)
	if err != nil {
		return nil, err
	}

	if match {
		return &entity, nil
	}

	return nil, nil
}

func (fs *FileStorage) listCollectionFiles(collectionPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(collectionPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории и временные файлы
		if d.IsDir() || filepath.Ext(path) == ".tmp" {
			return nil
		}

		// Получаем относительное имя файла
		relPath, err := filepath.Rel(collectionPath, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	return files, err
}

// save с атомарной записью (БЕЗ МЬЮТЕКСА!)
func (fs *FileStorage) save(ctx context.Context, id string, docFolder string, data Entity) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context error: %w", err)
	}

	// Генерируем ID если нужно
	if id == "" {
		id = uuid.New().String()
	}
	data["_id"] = id

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
func (fs *FileStorage) load(ctx context.Context, docFolder, docPath string, result *Entity) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	path := fs.getPath(docFolder, docPath)

	// Если файла нет - возвращаем nil без ошибки
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewNotFoundError("no found")
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

// delete удаляет файл по ID из указанной коллекции
func (fs *FileStorage) delete(ctx context.Context, docFolder string, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	// Формируем путь к файлу
	path := fs.getPath(docFolder, id)
	logrus.Debugf("deleting path %s", path)

	// Проверяем существование файла
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewNotFoundError(fmt.Sprintf("no file %s", path))
	}

	// Удаляем файл
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
