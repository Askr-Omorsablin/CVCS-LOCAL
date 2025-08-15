package api

import "main/core"

// GenericRequest 是所有请求的基础结构
type GenericRequest struct {
	Positions map[string]interface{} `json:"positions" binding:"required"`
	Content   map[string]interface{} `json:"content"` // content 可以是可选的
}

// === 初始化 Codebase ===
type InitCodebasePositions struct {
	// 初始化时通常不需要位置信息，但可以保留以备将来使用（如租户ID）
}
type InitCodebaseContent struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Branch      string `json:"branch" binding:"required"`
}
type InitCodebaseRequest struct {
	Positions InitCodebasePositions `json:"positions" binding:"required"`
	Content   InitCodebaseContent   `json:"content" binding:"required"`
}

// === 创建快照 ===
type CreateSnapshotPositions struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
}

type BranchFrom struct {
	Branch  string `json:"branch" binding:"required"`
	Version string `json:"version" binding:"required"`
}

type CreateSnapshotContent struct {
	CodebasePath string      `json:"codebase_path" binding:"required"`
	Version      string      `json:"version"`
	Branch       string      `json:"branch"`
	Message      string      `json:"message,omitempty"`
	BranchFrom   *BranchFrom `json:"branch_from,omitempty"` // 新增：分支来源
	AutoLinkage  bool        `json:"auto_linkage"`          // 新增：是否自动建立血缘关系，默认true
}
type CreateSnapshotRequest struct {
	Positions CreateSnapshotPositions `json:"positions" binding:"required"`
	Content   CreateSnapshotContent   `json:"content" binding:"required"`
}

// === 获取归档 ===
type GetArchivePositions struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
}
type GetArchiveContent struct {
	Branch  string `json:"branch" binding:"required"`
	Version string `json:"version" binding:"required"`
}
type GetArchiveRequest struct {
	Positions GetArchivePositions `json:"positions" binding:"required"`
	Content   GetArchiveContent   `json:"content" binding:"required"`
}

// === 获取单个文件 ===
type GetFilePositions struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
}
type GetFileContent struct {
	Branch  string `json:"branch" binding:"required"`
	Version string `json:"version" binding:"required"`
	Path    string `json:"path" binding:"required"`
}
type GetFileRequest struct {
	Positions GetFilePositions `json:"positions" binding:"required"`
	Content   GetFileContent   `json:"content" binding:"required"`
}

// === 删除 Codebase ===
type DeleteCodebasePositions struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
}
type DeleteCodebaseRequest struct {
	Positions DeleteCodebasePositions `json:"positions" binding:"required"`
}

// === 通用响应 ===
type SnapshotResponse struct {
	Codebase   *core.Codebase           `json:"codebase"`
	Version    *core.Version            `json:"version"`
	FileTree   *core.FileTree           `json:"file_tree"`
	VersionMap *core.VersionMapResponse `json:"version_map,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// === 获取版本历史 ===
type GetVersionMapPositions struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
}
type GetVersionMapRequest struct {
	Positions GetVersionMapPositions `json:"positions" binding:"required"`
}

// === 创建版本链接 ===
type VersionIdentifier struct {
	Branch  string `json:"branch" binding:"required"`
	Version string `json:"version" binding:"required"`
}

type CreateVersionLinkContent struct {
	ChildVersion  VersionIdentifier `json:"child_version" binding:"required"`
	ParentVersion VersionIdentifier `json:"parent_version" binding:"required"`
}

type CreateVersionLinkRequest struct {
	Positions struct {
		CodebaseID string `json:"codebase_id" binding:"required"`
	} `json:"positions" binding:"required"`
	Content CreateVersionLinkContent `json:"content" binding:"required"`
}

// === 版本血缘关系结构 ===

// LinkageType 血缘关系类型
type LinkageType string

const (
	LinkageTypeSequential LinkageType = "sequential"  // 同分支时间序列血缘
	LinkageTypeBranchFrom LinkageType = "branch_from" // 跨分支创建血缘
)

// === 配置管理 ===
type UpdateStoragePathContent struct {
	Path string `json:"path" binding:"required"`
}

type UpdateStoragePathRequest struct {
	Content UpdateStoragePathContent `json:"content" binding:"required"`
}

// VersionEdge 代表图中的一条边
type VersionEdge struct {
	From        string      `json:"from"`         // parent_version_id
	To          string      `json:"to"`           // child_version_id
	LinkageType LinkageType `json:"linkage_type"` // 血缘关系类型
}
