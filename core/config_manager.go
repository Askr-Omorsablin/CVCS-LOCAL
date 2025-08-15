package core

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig defines application configuration parameters.
type AppConfig struct {
	StoragePath string `json:"storage_path"`
}

var (
	globalConfig *AppConfig
	configOnce   sync.Once
	configMu     sync.RWMutex
)

const (
	defaultStoragePath = "./cvcs_data"
	configDirName      = "cvcs"
	configFileName     = "config.json"
)

// getConfigFilePath determines the absolute path of the configuration file.
func getConfigFilePath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Println("Warning: Unable to get user config directory, will use current directory")
		return configFileName, nil
	}
	return filepath.Join(userConfigDir, configDirName, configFileName), nil
}

// LoadConfig loads configuration from user config directory.
// If file doesn't exist, initializes with default values and saves.
func LoadConfig() {
	configOnce.Do(func() {
		path, err := getConfigFilePath()
		if err != nil {
			// If unable to determine path at startup, this is a fatal error
			log.Fatalf("Unable to determine config file path: %v", err)
		}

		// Initialize default configuration
		globalConfig = &AppConfig{StoragePath: defaultStoragePath}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Config file not found, will create default config at %s", path)
				// Try to save default config, continue with default values even if it fails
				if err := saveCurrentConfig(); err != nil {
					log.Printf("Warning: Failed to save default config: %v", err)
				}
				return // Use default configuration
			}
			// Other read errors, continue with default values
			log.Printf("Warning: Failed to read config file %s: %v", path, err)
			return
		}

		// Parse configuration file
		var loadedConfig AppConfig
		if err := json.Unmarshal(data, &loadedConfig); err != nil {
			log.Printf("Warning: Failed to parse config file, will use default values: %v", err)
			return // Use default configuration
		}
		globalConfig = &loadedConfig
	})
}

// saveCurrentConfig saves current in-memory configuration to file.
func saveCurrentConfig() error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// GetConfig returns a copy of current application configuration.
func GetConfig() AppConfig {
	LoadConfig() // Ensure loaded
	configMu.RLock()
	defer configMu.RUnlock()
	return *globalConfig
}

// UpdateConfig updates configuration in memory and persists it to file.
func UpdateConfig(newConfig AppConfig) error {
	LoadConfig() // Ensure initialized
	configMu.Lock()
	*globalConfig = newConfig
	configMu.Unlock()
	return saveCurrentConfig()
}
