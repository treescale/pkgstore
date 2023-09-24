package main

import (
	"github.com/alin-io/pkgproxy/config"
	_ "github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/alin-io/pkgproxy/router"
	"github.com/alin-io/pkgproxy/storage"
)

func main() {
	var storageBackend storage.BaseStorageBackend
	if config.Get().Storage.ActiveBackend == config.StorageS3 {
		storageBackend = storage.NewS3Backend()
	} else {
		panic("Unknown storage backend")
	}

	// Sync Models with the DB
	models.SyncModels()

	r := router.SetupGinServer(storageBackend)

	err := r.Run(config.Get().ListenAddress)
	if err != nil {
		panic(err)
	}
}
