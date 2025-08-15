package core

import "sync"

// DynamicStorage is a thread-safe wrapper for the Storage interface
// that allows for hot-swapping the underlying implementation.
type DynamicStorage struct {
	mu    sync.RWMutex
	store Storage
}

func NewDynamicStorage(initialStore Storage) *DynamicStorage {
	return &DynamicStorage{store: initialStore}
}

// Swap replaces the underlying storage implementation.
func (d *DynamicStorage) Swap(newStore Storage) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.store = newStore
}

// PutObject forwards the call to the underlying implementation.
func (d *DynamicStorage) PutObject(objectName string, data []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.store.PutObject(objectName, data)
}

// GetObject forwards the call to the underlying implementation.
func (d *DynamicStorage) GetObject(objectName string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.store.GetObject(objectName)
}

// DeleteObjectsWithPrefix forwards the call to the underlying implementation.
func (d *DynamicStorage) DeleteObjectsWithPrefix(prefix string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.store.DeleteObjectsWithPrefix(prefix)
}
