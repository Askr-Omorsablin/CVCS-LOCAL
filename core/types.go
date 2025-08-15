package core

import "time"

// Codebase 代码库元数据
type Codebase struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Branch      string    `json:"branch"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"` // 添加这行
}

// Version 版本快照信息
type Version struct {
	ID         string    `json:"id"`
	CodebaseID string    `json:"codebase_id"`
	Version    string    `json:"version"`
	Branch     string    `json:"branch"`
	TreeID     string    `json:"tree_id"`
	Message    string    `json:"message,omitempty"` // 新增版本信息
	CreatedAt  time.Time `json:"created_at"`
	Stats      struct {
		TotalFiles       int     `json:"total_files"`
		TotalSize        int64   `json:"total_size"`
		CompressedSize   int64   `json:"compressed_size"`
		CompressionRatio float64 `json:"compression_ratio"`
	} `json:"stats"`
}

// VersionStats 版本统计信息
type VersionStats struct {
	TotalFiles       int     `json:"total_files"`
	TotalSize        int64   `json:"total_size"`
	CompressedSize   int64   `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// FileTree 文件树详情
type FileTree struct {
	TreeID      string    `json:"tree_id"`
	VersionID   string    `json:"version_id"`
	Files       []File    `json:"files"`
	GeneratedAt time.Time `json:"generated_at"`
}

// File 文件信息
type File struct {
	Path           string `json:"path"`
	Hash           string `json:"hash"`
	Size           int64  `json:"size"`
	CompressedSize int64  `json:"compressed_size"`
	StorageKey     string `json:"storage_key"`
	Type           string `json:"type"`
}

// SnapshotRequest API请求结构
type SnapshotRequest struct {
	CodebasePath string `json:"codebase_path"`
	CodebaseName string `json:"codebase_name"`
	Branch       string `json:"branch"`
	Version      string `json:"version"` // 可扩展从请求获取版本号
}

// SnapshotResponse API响应结构
type SnapshotResponse struct {
	Codebase   *Codebase           `json:"codebase"`
	Version    *Version            `json:"version"`
	FileTree   *FileTree           `json:"file_tree"`
	VersionMap *VersionMapResponse `json:"version_map,omitempty"`
}

// === 版本历史图谱结构 ===

// VersionNode 代表图中的一个版本节点
type VersionNode struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Branch    string    `json:"branch"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Stats     struct {
		TotalFiles       int     `json:"total_files"`
		TotalSize        int64   `json:"total_size"`
		CompressedSize   int64   `json:"compressed_size"`
		CompressionRatio float64 `json:"compression_ratio"`
	} `json:"stats"`
}

// === 版本血缘关系结构 ===

// LinkageType 血缘关系类型
type LinkageType string

const (
	LinkageTypeSequential LinkageType = "sequential"  // 同分支时间序列血缘
	LinkageTypeBranchFrom LinkageType = "branch_from" // 跨分支创建血缘
)

// VersionEdge 代表图中的一条边
type VersionEdge struct {
	From        string      `json:"from"`         // parent_version_id
	To          string      `json:"to"`           // child_version_id
	LinkageType LinkageType `json:"linkage_type"` // 血缘关系类型
}

// VersionMapResponse 是 /map API 的响应体
type VersionMapResponse struct {
	CodebaseID string            `json:"codebase_id"`
	Nodes      []VersionNode     `json:"nodes"`
	Edges      []VersionEdge     `json:"edges"`
	Refs       map[string]string `json:"refs"`
}
