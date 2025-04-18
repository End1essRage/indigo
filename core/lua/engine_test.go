package lua_test

//check asyncs
//check panics

import (
	"os"
	"path/filepath"
	"testing"

	l "github.com/end1essrage/indigo-core/lua"
)

func Test_LoadScripts(t *testing.T) {
	// Создаем временную директорию с тестовыми скриптами
	tmpDir := filepath.Join(os.TempDir(), "indigo")
	if err := os.Mkdir(tmpDir, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	// Создаем структуру директорий и файлов:
	// tmpDir/
	//   root_script.lua
	//   subdir1/
	//     sub1_script.lua
	//     subdir2/
	//       sub2_script.lua

	// Корневой скрипт
	rootScript := "root script content"
	if err := os.WriteFile(filepath.Join(tmpDir, "root_script.lua"), []byte(rootScript), 0644); err != nil {
		t.Fatal(err)
	}

	// Поддиректория 1
	subDir1 := filepath.Join(tmpDir, "subdir1")
	if err := os.Mkdir(subDir1, 0755); err != nil {
		t.Fatal(err)
	}

	// Скрипт в поддиректории 1
	sub1Script := "subdir1 script content"
	if err := os.WriteFile(filepath.Join(subDir1, "sub1_script.lua"), []byte(sub1Script), 0644); err != nil {
		t.Fatal(err)
	}

	// Поддиректория 2 (вложенная в subdir1)
	subDir2 := filepath.Join(subDir1, "subdir2")
	if err := os.Mkdir(subDir2, 0755); err != nil {
		t.Fatal(err)
	}

	// Скрипт в поддиректории 2
	sub2Script := "subdir2 script content"
	if err := os.WriteFile(filepath.Join(subDir2, "sub2_script.lua"), []byte(sub2Script), 0644); err != nil {
		t.Fatal(err)
	}

	// Загружаем скрипты
	scripts, err := l.LoadScripts(tmpDir)
	if err != nil {
		t.Fatalf("LoadScripts failed: %v", err)
	}

	// Проверяем ожидаемые ключи
	testCases := []struct {
		key      string
		expected string
	}{
		{"root_script.lua", rootScript},
		{filepath.Join("subdir1", "sub1_script.lua"), sub1Script},
		{filepath.Join("subdir1", "subdir2", "sub2_script.lua"), sub2Script},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			content, ok := scripts[tc.key]
			if !ok {
				t.Errorf("expected key %q not found in scripts map", tc.key)
				return
			}

			if string(content) != tc.expected {
				t.Errorf("for key %q expected content %q, got %q",
					tc.key, tc.expected, string(content))
			}
		})
	}

	// Проверяем общее количество загруженных скриптов
	if len(scripts) != len(testCases) {
		t.Errorf("expected %d scripts, got %d", len(testCases), len(scripts))
	}
}

func Test_LoadScripts_EdgeCases(t *testing.T) {
	// Создаем временную директорию с тестовыми скриптами
	tmpDir := filepath.Join(os.TempDir(), "indigo")
	if err := os.Mkdir(tmpDir, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("empty directory", func(t *testing.T) {
		scripts, err := l.LoadScripts(tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(scripts) != 0 {
			t.Errorf("expected empty map, got %d items", len(scripts))
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := l.LoadScripts("/nonexistent/path")
		if err == nil {
			t.Error("expected error for non-existent directory, got nil")
		}
	})

	t.Run("directory with non-lua files", func(t *testing.T) {
		// Создаем файл без расширения .lua
		if err := os.WriteFile(filepath.Join(tmpDir, "not_a_script.txt"), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		scripts, err := l.LoadScripts(tmpDir)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(scripts) != 0 {
			t.Errorf("expected 0 file, got %d", len(scripts))
		}
		// должен фильтровать
		if _, ok := scripts["not_a_script.txt"]; ok {
			t.Error("file not_a_script.txt found in scripts map")
		}
	})
}
