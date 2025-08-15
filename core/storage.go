package core

// Storage 定义了对象存储操作的接口，
// 抽象了底层实现（例如，本地磁盘、OSS）。
type Storage interface {
	PutObject(objectName string, data []byte) error
	GetObject(objectName string) ([]byte, error)
	DeleteObjectsWithPrefix(prefix string) error
}
