package api

import (
	"encoding/json"
	"fmt"
	"main/calculate"
	"net/http"
	"os"
	"strings"

	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type SnapshotHandler struct {
	uploadService *calculate.UploadService
}

func NewSnapshotHandler() *SnapshotHandler {
	return &SnapshotHandler{
		uploadService: calculate.NewUploadService(),
	}
}

func (h *SnapshotHandler) CreateSnapshot(c *gin.Context) {
	// 1. Parse multipart/form-data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart/form-data: " + err.Error()})
		return
	}

	// 2. Parse metadata
	metadataValues := form.Value["metadata"]
	if len(metadataValues) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'metadata' field"})
		return
	}

	var req CreateSnapshotRequest
	if err := json.Unmarshal([]byte(metadataValues[0]), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata JSON format: " + err.Error()})
		return
	}

	// 3. Get files
	files := make(map[string]*multipart.FileHeader)
	for key, fileHeaders := range form.File {
		if len(fileHeaders) > 0 {
			files[key] = fileHeaders[0]
		}
	}

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files found in request"})
		return
	}

	// 4. Set default values
	branch := req.Content.Branch
	if branch == "" {
		branch = "main"
	}
	version := req.Content.Version
	if version == "" {
		version = "v1"
	}
	message := req.Content.Message
	autoLinkage := req.Content.AutoLinkage

	// 5. Convert BranchFrom parameter
	var branchFrom *calculate.BranchFrom
	if req.Content.BranchFrom != nil {
		branchFrom = &calculate.BranchFrom{
			Branch:  req.Content.BranchFrom.Branch,
			Version: req.Content.BranchFrom.Version,
		}
	}

	// 6. Call refactored service
	resp, err := h.uploadService.ProcessSnapshot(
		req.Positions.CodebaseID,
		version,
		branch,
		message,
		files,
		branchFrom,  // Pass branch source information
		autoLinkage, // Pass automatic lineage flag
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type InitHandler struct {
	service *calculate.InitService
}

func NewInitHandler() *InitHandler {
	return &InitHandler{
		service: calculate.NewInitService(),
	}
}

func (h *InitHandler) Initialize(c *gin.Context) {
	var req InitCodebaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	codebase, err := h.service.InitializeCodebase(req.Content.Name, req.Content.Description, req.Content.Branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, codebase)
}

// ArchiveHandler handles archive download requests
type ArchiveHandler struct {
	service *calculate.ArchiveService
}

func NewArchiveHandler() *ArchiveHandler {
	return &ArchiveHandler{
		service: calculate.NewArchiveService(),
	}
}

func (h *ArchiveHandler) GetCodebaseArchive(c *gin.Context) {
	var req GetArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	codebaseName, err := h.service.GetCodebaseName(req.Positions.CodebaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Specified codebase not found"})
		return
	}

	zipPath, err := h.service.CreateArchiveForVersion(req.Positions.CodebaseID, req.Content.Branch, req.Content.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(zipPath)

	archiveFilename := fmt.Sprintf("%s-%s-%s.zip", codebaseName, req.Content.Branch, req.Content.Version)
	c.Header("Content-Disposition", "attachment; filename="+archiveFilename)
	c.File(zipPath)
}

func (h *ArchiveHandler) GetSingleFile(c *gin.Context) {
	var req GetFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	fileContent, fileName, err := h.service.GetSingleFile(req.Positions.CodebaseID, req.Content.Branch, req.Content.Version, req.Content.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// DeleteHandler handles delete requests
type DeleteHandler struct {
	service *calculate.DeleteService
}

func NewDeleteHandler() *DeleteHandler {
	return &DeleteHandler{
		service: calculate.NewDeleteService(),
	}
}

func (h *DeleteHandler) DeleteCodebase(c *gin.Context) {
	var req DeleteCodebaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	if err := h.service.DeleteCodebase(req.Positions.CodebaseID); err != nil {
		// Distinguish between "not found" and "other internal errors"
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "找不到") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error occurred while deleting codebase: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Codebase %s has been successfully deleted", req.Positions.CodebaseID)})
}

// ConfigHandler handles configuration requests
type ConfigHandler struct {
	service *calculate.ConfigService
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		service: calculate.NewConfigService(),
	}
}

func (h *ConfigHandler) UpdateStoragePath(c *gin.Context) {
	var req UpdateStoragePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	if err := h.service.SetStoragePath(req.Content.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Storage path updated successfully. Configuration is now in effect.",
		"path":    req.Content.Path,
	})
}
