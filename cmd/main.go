package main

import (
	"github.com/alin-io/pkgproxy/config"
	_ "github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/alin-io/pkgproxy/router"
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(gin.ErrorLogger())

	// Route Services
	r.GET("/", services.HealthCheckHandler)

	var storageBackend storage.BaseStorageBackend
	if config.Get().Storage.ActiveBackend == config.StorageS3 {
		storageBackend = storage.NewS3Backend()
	} else {
		panic("Unknown storage backend")
	}

	// Sync Models with the DB
	models.SyncModels()

	// NPM
	router.NPMRoutes(r, storageBackend)

	r.Run(config.Get().ListenAddress) // listen and serve on
}
