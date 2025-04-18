package helpers

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Spy struct {
	watcher  *fsnotify.Watcher
	Data     map[string][]byte // Путь -> содержимое файла
	dataLock sync.RWMutex
	rootPath string
}

func NewSpy(rootPath string) (*Spy, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &Spy{
		watcher:  watcher,
		Data:     make(map[string][]byte),
		rootPath: rootPath,
	}

	// Первоначальная загрузка файлов
	if err := filepath.WalkDir(rootPath, fw.walkDir); err != nil {
		return nil, err
	}

	go fw.run()
	return fw, nil
}

func (fw *Spy) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		return fw.loadFile(path)
	}

	// Добавляем директорию в наблюдение
	if err := fw.watcher.Add(path); err != nil {
		log.Printf("Ошибка добавления в наблюдение: %s: %v", path, err)
	}
	return nil
}

func (fw *Spy) loadFile(fullPath string) error {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(fw.rootPath, fullPath)
	if err != nil {
		return err
	}

	fw.dataLock.Lock()
	defer fw.dataLock.Unlock()
	fw.Data[relPath] = content
	return nil
}

func (fw *Spy) run() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Ошибка watcher: %v", err)
		}
	}
}

func (fw *Spy) handleEvent(event fsnotify.Event) {
	relPath, err := filepath.Rel(fw.rootPath, event.Name)
	if err != nil {
		log.Printf("Ошибка получения относительного пути: %v", err)
		return
	}

	switch {
	case event.Has(fsnotify.Create):
		if isDir(event.Name) {
			fw.addWatch(event.Name)
		} else {
			fw.loadFile(event.Name)
		}

	case event.Has(fsnotify.Write):
		if !isDir(event.Name) {
			fw.loadFile(event.Name)
		}

	case event.Has(fsnotify.Remove):
		fw.deleteFile(relPath)

	case event.Has(fsnotify.Rename):
		fw.deleteFile(relPath)
	}
}

func (fw *Spy) addWatch(path string) {
	if err := fw.watcher.Add(path); err != nil {
		log.Printf("Ошибка добавления наблюдения: %s: %v", path, err)
	}

	filepath.WalkDir(path, func(subPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fw.loadFile(subPath)
		}
		return nil
	})
}

func (fw *Spy) deleteFile(relPath string) {
	fw.dataLock.Lock()
	defer fw.dataLock.Unlock()
	delete(fw.Data, relPath)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (fw *Spy) Close() error {
	return fw.watcher.Close()
}
