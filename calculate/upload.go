package calculate

import (
	"encoding/json"
	"fmt"
	"log"
	"main/core"
	"mime/multipart"
	"time"
)

type UploadService struct {
	historyService *HistoryService
}

func NewUploadService() *UploadService {
	return &UploadService{
		historyService: NewHistoryService(),
	}
}

// BranchFrom defines branch source information
type BranchFrom struct {
	Branch  string `json:"branch"`
	Version string `json:"version"`
}

func (s *UploadService) ProcessSnapshot(codebaseID, ver, branch, message string, files map[string]*multipart.FileHeader, branchFrom *BranchFrom, autoLinkage bool) (*core.SnapshotResponse, error) {
	provider := core.GetProvider()
	storage := core.GetStore()

	// 1. Verify codebase exists
	codebase, err := provider.GetCodebaseByID(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid codebase_id: %s, error: %w", codebaseID, err)
	}
	codebaseInfo := *codebase

	// 2. Create snapshot (pass storage interface)
	_, versionJSON, fileTreeJSON, err := CreateSnapshot(
		storage,
		files,
		codebaseInfo.Name,
		branch,
		ver,
		message,
	)
	if err != nil {
		log.Printf("Snapshot creation failed: %v", err)
		return nil, err
	}

	// 3. Parse JSON data
	var (
		version  core.Version
		fileTree core.FileTree
	)
	if err := json.Unmarshal(versionJSON, &version); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(fileTreeJSON, &fileTree); err != nil {
		return nil, err
	}

	// 4. Associate CodebaseID to new version
	version.CodebaseID = codebaseID

	// 5. Persist metadata
	if err := s.persistMetadata(provider, codebase, &version, &fileTree); err != nil {
		return nil, err
	}

	// 6. Automatically establish lineage relationships (if enabled)
	if autoLinkage {
		if err := s.establishLinkage(codebaseID, version.ID, branch, branchFrom); err != nil {
			// Lineage relationship establishment failure should not affect snapshot creation, just log the error
			log.Printf("Failed to establish lineage relationship (version ID: %s): %v", version.ID, err)
		}
	}

	// 7. Synchronously get updated version graph
	var versionMap core.VersionMapResponse
	historyJSON, err := s.historyService.GetVersionMap(codebaseID)
	if err != nil {
		// Even if getting graph fails, should not interrupt snapshot creation process, just log error
		log.Printf("Unable to get version graph after creating snapshot (codebaseID: %s): %v", codebaseID, err)
	} else if historyJSON != nil {
		if err := json.Unmarshal(historyJSON, &versionMap); err != nil {
			log.Printf("Unable to parse version graph JSON (codebaseID: %s): %v", codebaseID, err)
		}
	}

	// Update codebaseInfo's UpdatedAt field
	codebaseInfo.UpdatedAt = time.Now()

	return &core.SnapshotResponse{
		Codebase:   &codebaseInfo,
		Version:    &version,
		FileTree:   &fileTree,
		VersionMap: &versionMap,
	}, nil
}

func (s *UploadService) persistMetadata(provider core.DataProvider, codebase *core.Codebase, version *core.Version, fileTree *core.FileTree) error {
	// Update codebase's updated_at field
	if err := provider.UpdateCodebaseTimestamp(codebase.ID, time.Now()); err != nil {
		return fmt.Errorf("failed to update codebase: %w", err)
	}

	// Save version and file indexes
	if err := provider.CreateVersion(version, fileTree.Files); err != nil {
		return fmt.Errorf("failed to insert version and file indexes: %w", err)
	}
	return nil
}

// establishLinkage establishes version lineage relationships
func (s *UploadService) establishLinkage(codebaseID, versionID, branch string, branchFrom *BranchFrom) error {
	if branchFrom != nil {
		// Cross-branch lineage relationship: create from specified source version
		log.Printf("Establishing cross-branch lineage relationship: %s/%s -> current version", branchFrom.Branch, branchFrom.Version)
		return s.historyService.AutoCreateBranchFromLink(codebaseID, versionID, branch, VersionIdentifier{
			Branch:  branchFrom.Branch,
			Version: branchFrom.Version,
		})
	}

	// Automatically detect link type
	isNew, err := s.historyService.isNewBranch(codebaseID, branch, versionID)
	if err != nil {
		return fmt.Errorf("failed to check if new branch: %w", err)
	}

	if isNew {
		// New branch: create link from main branch
		log.Printf("Detected new branch '%s', attempting to create link from main branch", branch)
		return s.historyService.AutoCreateLinkForNewBranch(codebaseID, versionID, branch)
	} else {
		// Existing branch: create same-branch sequential link
		log.Printf("Creating sequential link on existing branch '%s'", branch)
		return s.historyService.AutoCreateSequentialLink(codebaseID, branch, versionID)
	}
}
