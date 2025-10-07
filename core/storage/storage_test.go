package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// setupTestStorage создает временное хранилище для тестов
func setupTestStorage(t *testing.T) (*FileStorage, func()) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("test_storage_%d", time.Now().UnixNano()))

	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return storage, cleanup
}

// TestEmpty пустой тест для проверки что все компилируется
func TestEmpty(t *testing.T) {
	t.Log("Empty test passed")
}

// TestNewFileStorage проверяет создание хранилища
func TestNewFileStorage(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("test_storage_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tempDir)

	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Storage is nil")
	}

	// Проверяем что директория создана
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatal("Base directory was not created")
	}
}

// TestCreate базовый тест создания сущности
func TestCreate(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	entity := Entity{
		"name": "test_entity",
		"data": "test_data",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if id == "" {
		t.Fatal("ID should not be empty")
	}

	t.Logf("Successfully created entity with ID: %s", id)
}

// TestGetById тест получения сущности по ID
func TestGetById(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	// Создаем сущность
	entity := Entity{
		"name": "test_entity",
		"data": "test_data",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Получаем сущность
	result, err := fs.GetById(ctx, collection, id)
	if err != nil {
		t.Fatalf("GetById failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Проверяем данные
	if result["name"] != "test_entity" {
		t.Errorf("Expected name 'test_entity', got '%v'", result["name"])
	}

	t.Logf("Successfully retrieved entity: %+v", result)
}

// TestUpdateById тест обновления сущности
func TestUpdateById(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	// Создаем сущность
	entity := Entity{
		"name": "original",
		"data": "original_data",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Обновляем сущность
	updatedEntity := Entity{
		"name": "updated",
		"data": "updated_data",
	}

	err = fs.UpdateById(ctx, collection, id, updatedEntity)
	if err != nil {
		t.Fatalf("UpdateById failed: %v", err)
	}

	// Проверяем что данные обновились
	result, err := fs.GetById(ctx, collection, id)
	if err != nil {
		t.Fatalf("GetById failed: %v", err)
	}

	if result["name"] != "updated" {
		t.Errorf("Expected name 'updated', got '%v'", result["name"])
	}

	t.Log("Successfully updated entity")
}

// TestDeleteById тест удаления сущности
func TestDeleteById(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	// Создаем сущность
	entity := Entity{
		"name": "to_delete",
		"data": "data",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Удаляем сущность
	err = fs.DeleteById(ctx, collection, id)
	if err != nil {
		t.Fatalf("DeleteById failed: %v", err)
	}

	// Проверяем что файл удален
	path := fs.getPath(collection, id)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("File should be deleted")
	}

	t.Log("Successfully deleted entity")
}

// TestContextCancellation проверяет отмену операций через контекст
func TestContextCancellation(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	collection := "test_collection"

	t.Run("Create with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Отменяем сразу

		entity := Entity{"name": "test"}
		_, err := fs.Create(ctx, collection, entity)

		if err == nil {
			t.Fatal("Expected error with cancelled context")
		}

		if !containsString(err.Error(), "context") {
			t.Errorf("Expected context error, got: %v", err)
		}
	})

	t.Run("GetById with cancelled context", func(t *testing.T) {
		// Создаем сущность
		ctx := context.Background()
		entity := Entity{"name": "test"}
		id, err := fs.Create(ctx, collection, entity)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Пытаемся получить с отмененным контекстом
		ctxCancelled, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = fs.GetById(ctxCancelled, collection, id)
		if err == nil {
			t.Fatal("Expected error with cancelled context")
		}
	})

	t.Run("Create with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ждем истечения таймаута

		entity := Entity{"name": "test"}
		_, err := fs.Create(ctx, collection, entity)

		if err == nil {
			t.Fatal("Expected timeout error")
		}
	})
}

// TestConcurrentCreates проверяет параллельное создание сущностей
func TestConcurrentCreates(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	ids := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			entity := Entity{
				"name": fmt.Sprintf("entity_%d", index),
				"data": fmt.Sprintf("data_%d", index),
			}

			id, err := fs.Create(ctx, collection, entity)
			if err != nil {
				errors <- err
				return
			}
			ids <- id
		}(i)
	}

	wg.Wait()
	close(errors)
	close(ids)

	// Проверяем ошибки
	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Got %d errors during concurrent creates", errorCount)
	}

	// Проверяем уникальность ID
	idSet := make(map[string]bool)
	for id := range ids {
		if idSet[id] {
			t.Fatalf("Duplicate ID found: %s", id)
		}
		idSet[id] = true
	}

	if len(idSet) != numGoroutines {
		t.Errorf("Expected %d unique IDs, got %d", numGoroutines, len(idSet))
	}

	t.Logf("Successfully created %d entities concurrently", len(idSet))
}

// TestConcurrentReads проверяет параллельное чтение
func TestConcurrentReads(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	// Создаем сущность для чтения
	entity := Entity{
		"name": "shared",
		"data": "shared_data",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	const numReaders = 100
	var wg sync.WaitGroup
	errors := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result, err := fs.GetById(ctx, collection, id)
			if err != nil {
				errors <- err
				return
			}

			if result["name"] != "shared" {
				errors <- fmt.Errorf("unexpected name: %v", result["name"])
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Got %d errors during concurrent reads", errorCount)
	}

	t.Logf("Successfully performed %d concurrent reads", numReaders)
}

// TestConcurrentReadWrite проверяет параллельные чтение и запись
func TestConcurrentReadWrite(t *testing.T) {
	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "test_collection"

	// Создаем начальную сущность
	entity := Entity{
		"name": "concurrent",
		"data": "initial",
	}

	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup

	// Запускаем читателей
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := fs.GetById(ctx, collection, id)
				if err != nil {
					t.Logf("Read error: %v", err)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Запускаем писателей
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				updatedEntity := Entity{
					"name": fmt.Sprintf("updated_%d_%d", index, j),
					"data": fmt.Sprintf("data_%d_%d", index, j),
				}
				err := fs.UpdateById(ctx, collection, id, updatedEntity)
				if err != nil {
					t.Logf("Write error: %v", err)
				}
				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent read/write test completed")
}

// TestRaceConditions специальный тест для запуска с -race
func TestRaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race test in short mode")
	}

	fs, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	collection := "race_test"

	// Создаем общую сущность
	entity := Entity{"name": "race_test", "data": "data"}
	id, err := fs.Create(ctx, collection, entity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup

	// Смешиваем операции
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			switch index % 3 {
			case 0: // Read
				fs.GetById(ctx, collection, id)
			case 1: // Write
				updated := Entity{"name": fmt.Sprintf("updated_%d", index)}
				fs.UpdateById(ctx, collection, id, updated)
			case 2: // Create new
				newEntity := Entity{"name": fmt.Sprintf("new_%d", index)}
				fs.Create(ctx, collection, newEntity)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Race condition test completed")
}

// Вспомогательная функция
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
