package main

import (
	"log"
	"main/api"
	"main/core"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== Service Starting (Dynamic Local File Mode) ===")

	// 1. Load configuration from user directory (e.g., config.json)
	core.LoadConfig()

	// 2. Initialize provider manager (it will use loaded configuration)
	//    Trigger sync.Once initialization by calling GetProvider
	_ = core.GetProvider()

	log.Printf("Current storage path: %s", core.GetConfig().StoragePath)

	// 3. Start web service
	gin.SetMode(gin.ReleaseMode)
	router := api.NewRouter()

	log.Println("Service ready, listening on :8080")
	log.Fatal(router.Run(":8080"))
}
