package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// 纯函数：读取压缩文件
func ReadCompressed(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

// 纯函数：解压数据
func DecompressData(compressed []byte) ([]byte, error) {
	// 使用zlib解压
	return decompressZlib(compressed)
}

// 纯函数：生成输出路径
func GenerateOutputPath(inputPath, outputDir string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	return filepath.Join(outputDir, base[:len(base)-len(ext)])
}

// 存储层（隔离副作用）
func SaveDecompressed(data []byte, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(outputPath, data, 0644)
}

// 组合函数
func DecompressFile(inputPath, outputDir string) (string, error) {
	compressed, err := ReadCompressed(inputPath)
	if err != nil {
		return "", err
	}

	data, err := DecompressData(compressed)
	if err != nil {
		return "", err
	}

	outputPath := GenerateOutputPath(inputPath, outputDir)
	if err := SaveDecompressed(data, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}
