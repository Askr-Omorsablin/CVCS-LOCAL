package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements Storage interface for local file system storage.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create local storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) PutObject(objectName string, data []byte) error {
	path := filepath.Join(s.basePath, objectName)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (s *LocalStorage) GetObject(objectName string) ([]byte, error) {
	path := filepath.Join(s.basePath, objectName)
	data, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("object not found: %s", objectName)
	}
	return data, err
}

func (s *LocalStorage) DeleteObjectsWithPrefix(prefix string) error {
	// For safety, treat prefix as directory relative to base path
	dirPath := filepath.Join(s.basePath, prefix)
	if strings.TrimSpace(prefix) == "" || !strings.HasPrefix(dirPath, s.basePath) {
		return fmt.Errorf("not allowed to delete root directory or use empty prefix")
	}
	return os.RemoveAll(dirPath)
}
