package calculate

import (
	"fmt"
	"main/core"
)

// DeleteService handles deletion logic
type DeleteService struct{}

func NewDeleteService() *DeleteService {
	return &DeleteService{}
}

// DeleteCodebase deletes a codebase and all its associated data
func (s *DeleteService) DeleteCodebase(codebaseID string) error {
	provider := core.GetProvider()
	storage := core.GetStore()

	// 1. Get codebase information from metadata to locate storage folder
	codebase, err := provider.GetCodebaseByID(codebaseID)
	if err != nil {
		return fmt.Errorf("codebase with ID %s not found: %w", codebaseID, err)
	}

	// 2. Delete all related files from object storage
	prefix := fmt.Sprintf("%s/", codebase.Name)
	if err := storage.DeleteObjectsWithPrefix(prefix); err != nil {
		// If storage deletion fails, terminate operation to avoid data inconsistency
		return fmt.Errorf("failed to delete files from storage: %w", err)
	}

	// 3. Delete codebase record from metadata
	if err := provider.DeleteCodebaseByID(codebaseID); err != nil {
		return fmt.Errorf("metadata deletion failed: %w", err)
	}

	return nil
}
