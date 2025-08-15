package core

import "time"

// DataProvider 定义了元数据持久化的接口，
// 抽象了数据库实现（例如，PostgreSQL、JSON文件）。
type DataProvider interface {
	// Codebase 操作
	CreateCodebase(codebase *Codebase) error
	GetCodebaseByID(id string) (*Codebase, error)
	DeleteCodebaseByID(id string) error
	UpdateCodebaseTimestamp(id string, t time.Time) error

	// Version 操作
	CreateVersion(version *Version, files []File) error
	GetVersion(codebaseID, branch, version string) (*Version, error)
	GetFileIndexesByTreeID(treeID string) ([]File, error)
	FindLatestVersionInBranch(codebaseID, branch, excludeVersionID string) (*Version, error)
	IsNewBranch(codebaseID, branch, excludeVersionID string) (bool, error)
	FindLatestVersionInMain(codebaseID string) (*Version, error)

	// History 和 Linkage 操作
	CreateVersionLink(codebaseID, childID, parentID, branch string, linkType LinkageType) error
	GetAllVersionsForMap(codebaseID string) ([]VersionNode, error)
	GetAllVersionEdgesForMap(codebaseID string) ([]VersionEdge, error)
	GetBranchHeadsForMap(codebaseID string) (map[string]string, error)

	// History Cache 操作
	GetHistoryCache(codebaseID string) ([]byte, error)
	UpdateHistoryCache(codebaseID string, data []byte) error
}
