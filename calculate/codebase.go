package calculate

import (
	"encoding/json"
	"fmt"
	"io"
	"main/core"
	"main/utils"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func CreateSnapshot(storage core.Storage, files map[string]*multipart.FileHeader, codebaseName, branch, version, message string) ([]byte, []byte, []byte, error) {
	// 1. 处理文件并创建版本
	versionID := uuid.NewString()
	treeID := uuid.NewString()

	processedFiles, stats, err := processFiles(storage, files, codebaseName)
	if err != nil {
		return nil, nil, nil, err
	}

	versionData := core.Version{
		ID:        versionID,
		Version:   version,
		Branch:    branch,
		Message:   message,
		TreeID:    treeID,
		CreatedAt: time.Now(),
		Stats:     stats,
	}

	// 2. 创建文件树
	fileTree := core.FileTree{
		TreeID:      treeID,
		VersionID:   versionID,
		Files:       processedFiles,
		GeneratedAt: time.Now(),
	}

	// 3. 序列化为JSON
	versionJSON, _ := json.MarshalIndent(versionData, "", "  ")
	fileTreeJSON, _ := json.MarshalIndent(fileTree, "", "  ")

	return nil, versionJSON, fileTreeJSON, nil
}

func processFiles(storage core.Storage, files map[string]*multipart.FileHeader, codebaseName string) ([]core.File, core.VersionStats, error) {
	var (
		processedFiles []core.File
		stats          core.VersionStats
		wg             sync.WaitGroup
		mu             sync.Mutex
		errChan        = make(chan error, len(files))
	)

	for relativePath, fileHeader := range files {
		wg.Add(1)
		go func(relPath string, header *multipart.FileHeader) {
			defer wg.Done()

			file, err := processFile(storage, relPath, header, codebaseName)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			processedFiles = append(processedFiles, file)
			stats.TotalFiles++
			stats.TotalSize += file.Size
			stats.CompressedSize += file.CompressedSize
			mu.Unlock()
		}(relativePath, fileHeader)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, core.VersionStats{}, err
		}
	}

	if stats.TotalSize > 0 {
		stats.CompressionRatio = float64(stats.CompressedSize) / float64(stats.TotalSize)
	}

	return processedFiles, stats, nil
}

func processFile(storage core.Storage, relativePath string, header *multipart.FileHeader, codebaseName string) (core.File, error) {
	// 1. 从文件头中读取文件内容
	file, err := header.Open()
	if err != nil {
		return core.File{}, fmt.Errorf("打开上传的文件流失败 %s: %w", relativePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return core.File{}, fmt.Errorf("读取上传的文件流失败 %s: %w", relativePath, err)
	}
	originalSize := int64(len(data))

	// 2. 根据扩展名判断是否为图片
	ext := strings.ToLower(filepath.Ext(relativePath))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff"}
	isImage := false
	for _, imgExt := range imageExts {
		if ext == imgExt {
			isImage = true
			break
		}
	}

	var contentToUpload []byte
	var compressedSize int64
	var fileType string

	// 3. 分类处理：图片不压缩，其他文件压缩
	if isImage {
		contentToUpload = data
		compressedSize = originalSize
		fileType = "image"
	} else {
		compressedData, err := utils.CompressData(data)
		if err != nil {
			return core.File{}, fmt.Errorf("压缩文件失败 %s: %w", relativePath, err)
		}
		contentToUpload = compressedData
		compressedSize = int64(len(compressedData))
		fileType = "other"
	}

	// 4. 公共处理流程: 计算哈希、生成Key、上传
	hash := utils.CalculateHash(data)
	// The storage key is now based on the content hash for deduplication and consistency.
	storageKey := fmt.Sprintf("%s/%s", codebaseName, hash)

	if err := storage.PutObject(storageKey, contentToUpload); err != nil {
		return core.File{}, fmt.Errorf("存储上传失败 %s: %w", relativePath, err)
	}

	// 5. 返回文件元数据
	return core.File{
		Path:           filepath.ToSlash(relativePath),
		Hash:           hash,
		Size:           originalSize,
		CompressedSize: compressedSize,
		StorageKey:     storageKey,
		Type:           fileType,
	}, nil
}
