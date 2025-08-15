package calculate

import (
	"main/core"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// SetStoragePath updates storage path configuration.
// It updates the configuration in memory, saves it to file, and triggers data/storage provider reinitialization.
func (s *ConfigService) SetStoragePath(newPath string) error {
	// 1. Get copy of current configuration to update field
	currentConfig := core.GetConfig()
	currentConfig.StoragePath = newPath

	// 2. Update configuration in memory and save to file
	if err := core.UpdateConfig(currentConfig); err != nil {
		return err
	}

	// 3. Trigger provider manager to reinitialize with new path
	//    (As requested, data migration is not handled here)
	return core.UpdateProviders(currentConfig)
}
