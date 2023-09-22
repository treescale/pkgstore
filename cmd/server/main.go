package main

import (
	"github.com/alin-io/pkgproxy/config"
	_ "github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/middlewares"
	"github.com/alin-io/pkgproxy/models"
	"github.com/alin-io/pkgproxy/router"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(middlewares.AuthMiddleware)

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

	err := r.Run(config.Get().ListenAddress)
	if err != nil {
		panic(err)
	}
}
