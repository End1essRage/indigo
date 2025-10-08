package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// GetSequential - последовательная обработка без горутин
func (fs *FileStorage) GetSequential(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	collectionPath := filepath.Join(fs.basePath, collection)

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return []Entity{}, nil
	}

	files, err := fs.listCollectionFiles(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return []Entity{}, nil
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

// GetWorkerPoolOptimized - оптимизированная версия worker pool
func (fs *FileStorage) GetWorkerPoolOptimized(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	collectionPath := filepath.Join(fs.basePath, collection)

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return []Entity{}, nil
	}

	files, err := fs.listCollectionFiles(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return []Entity{}, nil
	}

	// Если файлов мало - используем последовательную обработку
	if len(files) < 50 {
		return fs.GetSequential(ctx, collection, count, query)
	}

	// Используем буферизованные каналы правильного размера
	type result struct {
		entity Entity
		index  int // для сохранения порядка
	}

	resultsChan := make(chan result)
	filesChan := make(chan struct {
		name  string
		index int
	})

	var wg sync.WaitGroup
	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// Запускаем workers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range filesChan {
			select {
			case <-ctxCancel.Done():
				return
			default:
			}

			entity, err := fs.loadAndFilter(ctx, collection, job.name, query)
			if err != nil {
				continue
			}

			if entity != nil {
				select {
				case resultsChan <- result{entity: *entity, index: job.index}:
				case <-ctxCancel.Done():
					return
				}
			}
		}
	}()

	// Отправляем задачи
	go func() {
		for i, file := range files {
			select {
			case <-ctxCancel.Done():
				close(filesChan)
				return
			case filesChan <- struct {
				name  string
				index int
			}{name: file, index: i}:
			}
		}
		close(filesChan)
	}()

	// Собираем результаты
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := make([]Entity, 0, count)
	for res := range resultsChan {
		results = append(results, res.entity)
		if count > 0 && len(results) >= count {
			cancel() // Останавливаем workers
			break
		}
	}

	return results, nil
}

// GetWorkerPoolBatched - обработка батчами для минимизации overhead
func (fs *FileStorage) GetWorkerPoolBatched(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	collectionPath := filepath.Join(fs.basePath, collection)

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		return []Entity{}, nil
	}

	files, err := fs.listCollectionFiles(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return []Entity{}, nil
	}

	// Для малого количества файлов - последовательно
	if len(files) < 100 {
		return fs.GetSequential(ctx, collection, count, query)
	}

	numWorkers := runtime.NumCPU()
	batchSize := (len(files) + numWorkers - 1) / numWorkers

	type batchResult struct {
		entities []Entity
		err      error
	}

	resultsChan := make(chan batchResult, numWorkers)
	var wg sync.WaitGroup

	// Обрабатываем батчами
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		wg.Add(1)

		go func(batch []string) {
			defer wg.Done()

			batchResults := make([]Entity, 0, len(batch))
			for _, fileName := range batch {
				select {
				case <-ctx.Done():
					resultsChan <- batchResult{err: ctx.Err()}
					return
				default:
				}

				entity, err := fs.loadAndFilter(ctx, collection, fileName, query)
				if err != nil {
					continue
				}

				if entity != nil {
					batchResults = append(batchResults, *entity)
				}
			}

			resultsChan <- batchResult{entities: batchResults}
		}(batch)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Собираем результаты
	results := make([]Entity, 0, count)
	for res := range resultsChan {
		if res.err != nil {
			return results, res.err
		}
		results = append(results, res.entities...)
		if count > 0 && len(results) >= count {
			results = results[:count]
			break
		}
	}

	return results, nil
}

// setupBenchmarkData создает тестовые данные для бенчмарка
func setupBenchmarkData(b *testing.B, fs *FileStorage, collection string, numFiles int) {
	b.Helper()

	ctx := context.Background()
	for i := 0; i < numFiles; i++ {
		entity := Entity{
			"id":     fmt.Sprintf("%d", i),
			"name":   fmt.Sprintf("Entity_%d", i),
			"value":  i,
			"status": "active",
		}
		_, err := fs.Create(ctx, collection, entity)
		if err != nil {
			b.Fatalf("Failed to create test data: %v", err)
		}
	}
}

// === Бенчмарки 100 файлов ===

func BenchmarkGet_Sequential_100Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 100)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetSequential(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolOptimized_100Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 100)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolOptimized(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolBatched_100Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 100)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolBatched(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// === Бенчмарки 1000 файлов ===

func BenchmarkGet_Sequential_1000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 1000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetSequential(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolOptimized_1000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 1000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolOptimized(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolBatched_1000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 1000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolBatched(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// === Бенчмарки 10000 файлов ===

func BenchmarkGet_Sequential_10000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 10000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetSequential(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolOptimized_10000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 10000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolOptimized(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolBatched_10000Files(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 10000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolBatched(ctx, "test", 0, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// === Бенчмарки с лимитом (early exit) ===

func BenchmarkGet_Sequential_WithLimit10(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 1000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetSequential(ctx, "test", 10, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet_WorkerPoolOptimized_WithLimit10(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir)
	setupBenchmarkData(b, fs, "test", 1000)

	ctx := context.Background()
	query := &Condition{Field: "status", Operator: "=", Value: "active"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetWorkerPoolOptimized(ctx, "test", 10, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}
