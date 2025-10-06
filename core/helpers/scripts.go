package helpers

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Scripts struct {
	Data     map[string][]byte // Путь -> содержимое файла
	rootPath string
}

func NewScripts(rootPath string) (*Scripts, error) {
	fw := &Scripts{
		Data:     make(map[string][]byte),
		rootPath: rootPath,
	}

	// Первоначальная загрузка файлов
	if err := filepath.WalkDir(rootPath, fw.walkDir); err != nil {
		return nil, err
	}

	return fw, nil
}

func (fw *Scripts) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		return fw.loadFile(path)
	}

	return nil
}

func (fw *Scripts) loadFile(fullPath string) error {

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(fw.rootPath, fullPath)
	if err != nil {
		return err
	}

	//отбрасываем .lua
	shards := strings.Split(relPath, "\\")
	path := strings.Join(shards, "/")
	key := strings.Split(path, ".")[0]

	fw.Data[key] = content
	return nil
}
