package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 纯函数：读取文件内容
func ReadFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

// 纯函数：计算文件哈希
func CalculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// 纯函数：压缩数据
func CompressData(data []byte) ([]byte, error) {
	// 使用zlib压缩
	// 注意：Go的compress/zlib默认压缩级别是DefaultCompression(6)
	return compressZlib(data)
}

// 存储层（隔离副作用）
func SaveCompressedData(compressed []byte, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(outputPath, compressed, 0644)
}

// 组合函数
func CompressFile(inputPath, outputDir string) (string, string, error) {
	data, err := ReadFile(inputPath)
	if err != nil {
		return "", "", err
	}

	compressed, err := CompressData(data)
	if err != nil {
		return "", "", err
	}

	fileHash := CalculateHash(data)
	outputPath := filepath.Join(outputDir, fileHash+".compressed")

	if err := SaveCompressedData(compressed, outputPath); err != nil {
		return "", "", err
	}

	return outputPath, fileHash, nil
}
