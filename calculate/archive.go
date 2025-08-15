package calculate

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"main/core"
	"main/utils"
	"os"
	"path/filepath"
	"sync"
)

// ArchiveService handles codebase archiving logic
type ArchiveService struct{}

func NewArchiveService() *ArchiveService {
	return &ArchiveService{}
}

// CreateArchiveForVersion creates a zip archive for the specified version
// Returns the path of the temporarily generated zip file
func (s *ArchiveService) CreateArchiveForVersion(codebaseID, branch, version string) (string, error) {
	files, err := s.getFilesForVersion(codebaseID, branch, version)
	if err != nil {
		return "", fmt.Errorf("unable to get file list: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files found: codebase %s, branch %s, version %s", codebaseID, branch, version)
	}

	reconstructionDir, err := os.MkdirTemp("", "codebase-reconstruction-*")
	if err != nil {
		return "", fmt.Errorf("unable to create temporary reconstruction directory: %w", err)
	}
	defer os.RemoveAll(reconstructionDir)
	log.Printf("Reconstructing codebase in temporary directory: %s", reconstructionDir)

	if err := s.reconstructFiles(files, reconstructionDir); err != nil {
		return "", fmt.Errorf("file reconstruction failed: %w", err)
	}

	zipPath, err := s.createZipArchive(reconstructionDir)
	if err != nil {
		return "", fmt.Errorf("zip archive creation failed: %w", err)
	}

	log.Printf("Archive file created successfully: %s", zipPath)
	return zipPath, nil
}

// GetCodebaseName gets codebase name from metadata
func (s *ArchiveService) GetCodebaseName(codebaseID string) (string, error) {
	provider := core.GetProvider()
	codebase, err := provider.GetCodebaseByID(codebaseID)
	if err != nil {
		return "", err
	}
	return codebase.Name, nil
}

func (s *ArchiveService) getFilesForVersion(codebaseID, branch, version string) ([]core.File, error) {
	provider := core.GetProvider()
	v, err := provider.GetVersion(codebaseID, branch, version)
	if err != nil {
		return nil, fmt.Errorf("specified version not found: %w", err)
	}

	files, err := provider.GetFileIndexesByTreeID(v.TreeID)
	if err != nil {
		return nil, fmt.Errorf("file index query failed: %w", err)
	}
	return files, nil
}

func (s *ArchiveService) reconstructFiles(files []core.File, destDir string) error {
	storage := core.GetStore()
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(f core.File) {
			defer wg.Done()
			log.Printf("Processing file: %s (type: %s)", f.Path, f.Type)

			content, err := storage.GetObject(f.StorageKey)
			if err != nil {
				errChan <- fmt.Errorf("download %s failed: %w", f.StorageKey, err)
				return
			}

			if f.Type != "image" {
				content, err = utils.DecompressData(content)
				if err != nil {
					errChan <- fmt.Errorf("decompression %s failed: %w", f.Path, err)
					return
				}
			}

			destPath := filepath.Join(destDir, f.Path)
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				errChan <- fmt.Errorf("directory creation for %s failed: %w", f.Path, err)
				return
			}
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				errChan <- fmt.Errorf("file write %s failed: %w", f.Path, err)
				return
			}
		}(file)
	}

	wg.Wait()
	close(errChan)

	// Return the first error encountered
	return <-errChan
}

func (s *ArchiveService) createZipArchive(sourceDir string) (string, error) {
	zipFile, err := os.CreateTemp("", "codebase-archive-*.zip")
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipEntry, err := writer.Create(filepath.ToSlash(relPath))
		if err != nil {
			return err
		}

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(zipEntry, fileToZip)
		return err
	})

	return zipFile.Name(), err
}

// GetSingleFile downloads a single file
func (s *ArchiveService) GetSingleFile(codebaseID, branch, version, filePath string) ([]byte, string, error) {
	provider := core.GetProvider()
	storage := core.GetStore()

	// 1. Get version information
	v, err := provider.GetVersion(codebaseID, branch, version)
	if err != nil {
		return nil, "", fmt.Errorf("specified version not found: %w", err)
	}

	// 2. Find specific file from file index
	files, err := provider.GetFileIndexesByTreeID(v.TreeID)
	if err != nil {
		return nil, "", fmt.Errorf("file index for tree_id %s not found: %w", v.TreeID, err)
	}

	var targetFile *core.File
	for i := range files {
		if files[i].Path == filePath {
			targetFile = &files[i]
			break
		}
	}

	if targetFile == nil {
		return nil, "", fmt.Errorf("file '%s' not found in version %s", filePath, version)
	}

	// 3. Download file from storage
	content, err := storage.GetObject(targetFile.StorageKey)
	if err != nil {
		return nil, "", fmt.Errorf("file download failed: %w", err)
	}

	// 4. Decompress if not an image file
	if targetFile.Type != "image" {
		content, err = utils.DecompressData(content)
		if err != nil {
			return nil, "", fmt.Errorf("file decompression failed: %w", err)
		}
	}

	// Return file content and original filename
	fileName := filepath.Base(filePath)
	return content, fileName, nil
}
