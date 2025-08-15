package api

import (
	"main/calculate"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HistoryHandler handles version history related requests
type HistoryHandler struct {
	service *calculate.HistoryService
}

func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{
		service: calculate.NewHistoryService(),
	}
}

// GetVersionMap gets version history graph
func (h *HistoryHandler) GetVersionMap(c *gin.Context) {
	var req GetVersionMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	historyJSON, err := h.service.GetVersionMap(req.Positions.CodebaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", historyJSON)
}

// CreateVersionLink creates version link
func (h *HistoryHandler) CreateVersionLink(c *gin.Context) {
	var req CreateVersionLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body does not conform to specification: " + err.Error()})
		return
	}

	child := calculate.VersionIdentifier{
		Branch:  req.Content.ChildVersion.Branch,
		Version: req.Content.ChildVersion.Version,
	}
	parent := calculate.VersionIdentifier{
		Branch:  req.Content.ParentVersion.Branch,
		Version: req.Content.ParentVersion.Version,
	}

	if err := h.service.CreateVersionLink(req.Positions.CodebaseID, child, parent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version link created successfully"})
}
