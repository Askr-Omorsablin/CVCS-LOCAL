package api

import (

	// "main/core" //不再需要

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 1. 初始化所有处理器
	//    处理器现在自己管理其服务的生命周期
	initHandler := NewInitHandler()
	snapshotHandler := NewSnapshotHandler()
	archiveHandler := NewArchiveHandler()
	deleteHandler := NewDeleteHandler()
	historyHandler := NewHistoryHandler()
	configHandler := NewConfigHandler()

	api := r.Group("/api/v1")
	{
		// 所有端点统一为 POST
		api.POST("/codebases/init", initHandler.Initialize)
		api.POST("/codebases/snapshots/create", snapshotHandler.CreateSnapshot)
		api.POST("/codebases/archive/get", archiveHandler.GetCodebaseArchive)
		api.POST("/codebases/file/get", archiveHandler.GetSingleFile)
		api.POST("/codebases/delete", deleteHandler.DeleteCodebase)

		// 历史相关API
		api.POST("/codebases/map/get", historyHandler.GetVersionMap)
		api.POST("/codebases/map/link", historyHandler.CreateVersionLink)

		// 配置相关API
		api.POST("/config/storage/update", configHandler.UpdateStoragePath)
	}

	return r
}
