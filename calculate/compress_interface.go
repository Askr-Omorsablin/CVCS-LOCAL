package calculate

import (
	"main/utils"
	"path/filepath"
)

// CompressFile 暴露给外部的压缩接口
func CompressFile(inputPath, outputDir string) (string, string, error) {
	// 标准化路径处理
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", "", err
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", "", err
	}

	// 调用utils层的压缩功能
	return utils.CompressFile(absInputPath, absOutputDir)
}
