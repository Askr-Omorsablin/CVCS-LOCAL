package calculate

import (
	"encoding/json"
	"fmt"
	"main/core"
)

// VersionIdentifier defines the information needed to locate a version
type VersionIdentifier struct {
	Branch  string
	Version string
}

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

// GetVersionMap attempts to get version history graph from cache, builds it if failed.
// Returns raw JSON bytes that can be used directly for API response.
func (s *HistoryService) GetVersionMap(codebaseID string) ([]byte, error) {
	provider := core.GetProvider()
	historyJSON, err := provider.GetHistoryCache(codebaseID)
	if err == nil {
		return historyJSON, nil
	}

	// Cache miss or error, need to rebuild
	_, err = s.RebuildHistoryCache(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to build history cache: %w", err)
	}

	// Try to read from cache again
	historyJSON, err = provider.GetHistoryCache(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to read newly built cache: %w", err)
	}

	return historyJSON, nil
}

// CreateVersionLink creates parent-child link between two versions.
func (s *HistoryService) CreateVersionLink(codebaseID string, child, parent VersionIdentifier) error {
	return s.CreateVersionLinkWithType(codebaseID, child, parent, core.LinkageTypeSequential)
}

// CreateVersionLinkWithType creates version link of specified type
func (s *HistoryService) CreateVersionLinkWithType(codebaseID string, child, parent VersionIdentifier, linkageType core.LinkageType) error {
	provider := core.GetProvider()
	childVersion, err := provider.GetVersion(codebaseID, child.Branch, child.Version)
	if err != nil {
		return fmt.Errorf("child version '%s' (branch: %s) not found: %w", child.Version, child.Branch, err)
	}

	parentVersion, err := provider.GetVersion(codebaseID, parent.Branch, parent.Version)
	if err != nil {
		return fmt.Errorf("parent version '%s' (branch: %s) not found: %w", parent.Version, parent.Branch, err)
	}

	err = provider.CreateVersionLink(codebaseID, childVersion.ID, parentVersion.ID, child.Branch, linkageType)
	if err != nil {
		return fmt.Errorf("version link insertion failed: %w", err)
	}

	// After success, asynchronously rebuild cache
	go s.RebuildHistoryCache(codebaseID)

	return nil
}

// AutoCreateSequentialLink automatically creates time-series lineage relationships for the same branch
func (s *HistoryService) AutoCreateSequentialLink(codebaseID, branch, currentVersionID string) error {
	provider := core.GetProvider()
	// Find the latest version in the same branch (excluding current version)
	parentVersion, err := provider.FindLatestVersionInBranch(codebaseID, branch, currentVersionID)
	if err != nil {
		return fmt.Errorf("failed to find parent version: %w", err)
	}
	if parentVersion == nil {
		// This is the first version in this branch, no need to establish lineage relationship
		return nil
	}

	// Create lineage relationship
	err = provider.CreateVersionLink(codebaseID, currentVersionID, parentVersion.ID, branch, core.LinkageTypeSequential)
	if err != nil {
		return fmt.Errorf("failed to create sequential lineage relationship: %w", err)
	}

	return nil
}

// AutoCreateBranchFromLink automatically creates lineage relationship for cross-branch
func (s *HistoryService) AutoCreateBranchFromLink(codebaseID, newVersionID, childBranch string, branchFrom VersionIdentifier) error {
	provider := core.GetProvider()
	// Find source version ID
	parentVersion, err := provider.GetVersion(codebaseID, branchFrom.Branch, branchFrom.Version)
	if err != nil {
		return fmt.Errorf("source version %s/%s not found: %w", branchFrom.Branch, branchFrom.Version, err)
	}

	// Create lineage relationship
	err = provider.CreateVersionLink(codebaseID, newVersionID, parentVersion.ID, childBranch, core.LinkageTypeBranchFrom)
	if err != nil {
		return fmt.Errorf("failed to create branch lineage relationship: %w", err)
	}

	return nil
}

// RebuildHistoryCache rebuilds complete history graph for specified codebase and stores in cache.
// It now returns the built graph so callers can use it directly.
func (s *HistoryService) RebuildHistoryCache(codebaseID string) (*core.VersionMapResponse, error) {
	provider := core.GetProvider()
	nodes, err := provider.GetAllVersionsForMap(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("node query failed: %w", err)
	}

	edges, err := provider.GetAllVersionEdgesForMap(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("edge query failed: %w", err)
	}

	refs, err := provider.GetBranchHeadsForMap(codebaseID)
	if err != nil {
		return nil, fmt.Errorf("reference query failed: %w", err)
	}

	// Assemble response and serialize
	historyMap := core.VersionMapResponse{
		CodebaseID: codebaseID,
		Nodes:      nodes,
		Edges:      edges,
		Refs:       refs,
	}
	historyJSON, err := json.Marshal(historyMap)
	if err != nil {
		return nil, fmt.Errorf("history graph serialization failed: %w", err)
	}

	// Use UPSERT to update cache
	if err := provider.UpdateHistoryCache(codebaseID, historyJSON); err != nil {
		return nil, fmt.Errorf("cache update failed: %w", err)
	}

	return &historyMap, nil
}

// isNewBranch checks if a branch is completely new (no other versions except the current one being created)
func (s *HistoryService) isNewBranch(codebaseID, branch, currentVersionID string) (bool, error) {
	provider := core.GetProvider()
	return provider.IsNewBranch(codebaseID, branch, currentVersionID)
}

// AutoCreateLinkForNewBranch automatically creates link for a completely new branch, usually linking from the latest version of main branch
func (s *HistoryService) AutoCreateLinkForNewBranch(codebaseID, newVersionID, childBranch string) error {
	provider := core.GetProvider()
	// Find the latest version of main branch as parent version
	parentVersion, err := provider.FindLatestVersionInMain(codebaseID)
	if err != nil {
		return fmt.Errorf("failed to find parent version in main branch: %w", err)
	}

	if parentVersion == nil {
		// If main branch doesn't exist (or has no versions), this is the first version of an orphan branch, no linking needed
		return nil
	}

	// Create 'branch_from' type lineage relationship
	err = provider.CreateVersionLink(codebaseID, newVersionID, parentVersion.ID, childBranch, core.LinkageTypeBranchFrom)
	if err != nil {
		return fmt.Errorf("failed to create 'branch_from' link for new branch: %w", err)
	}

	return nil
}
