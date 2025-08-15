package core

import (
	"log"
	"path/filepath"
	"sync"
)

var (
	providerManager     *ProviderManager
	providerManagerOnce sync.Once
)

// ProviderManager holds dynamic instances of data and storage providers.
type ProviderManager struct {
	mu       sync.RWMutex
	provider DataProvider
	store    Storage
	config   AppConfig
}

// initProviderManager initializes the global providerManager singleton.
func initProviderManager() {
	providerManagerOnce.Do(func() {
		cfg := GetConfig()
		providerManager = &ProviderManager{config: cfg}
		if err := providerManager.reinitialize(); err != nil {
			log.Fatalf("Failed to initialize Provider Manager: %v", err)
		}
	})
}

// GetProvider returns the current DataProvider instance.
func GetProvider() DataProvider {
	initProviderManager() // Ensure initialized
	providerManager.mu.RLock()
	defer providerManager.mu.RUnlock()
	return providerManager.provider
}

// GetStore returns the current Storage instance.
func GetStore() Storage {
	initProviderManager() // Ensure initialized
	providerManager.mu.RLock()
	defer providerManager.mu.RUnlock()
	return providerManager.store
}

// reinitialize creates new provider instances based on current configuration.
func (pm *ProviderManager) reinitialize() error {
	dbPath := filepath.Join(pm.config.StoragePath, "db")
	storagePath := filepath.Join(pm.config.StoragePath, "oss")

	log.Printf("Reinitializing providers with new path: %s", pm.config.StoragePath)

	newStore, err := NewLocalStorage(storagePath)
	if err != nil {
		return err
	}

	newProvider, err := NewJSONFileProvider(dbPath)
	if err != nil {
		return err
	}

	pm.store = newStore
	pm.provider = newProvider
	return nil
}

// UpdateProviders reinitializes data and storage providers with new configuration.
func UpdateProviders(newConfig AppConfig) error {
	initProviderManager() // Ensure initialized
	providerManager.mu.Lock()
	defer providerManager.mu.Unlock()

	providerManager.config = newConfig
	return providerManager.reinitialize()
}
