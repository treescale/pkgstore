package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/config"
	"github.com/treescale/pkgstore/db"
	"github.com/treescale/pkgstore/models"
	"github.com/treescale/pkgstore/router"
	"github.com/treescale/pkgstore/services"
	"github.com/treescale/pkgstore/storage"
)

func main() {
	var storageBackend storage.BaseStorageBackend
	if config.Get().Storage.ActiveBackend == config.StorageS3 {
		storageBackend = storage.NewS3Backend()
	} else if config.Get().Storage.ActiveBackend == config.StorageFileSystem {
		storageBackend = storage.NewFileSystemBackend(config.Get().Storage.FileSystemRoot)
	} else {
		panic("Unknown storage backend")
	}

	// Initialize the DB connection
	db.InitDatabase()

	// Sync Models with the DB
	models.SyncModels()

	if len(os.Args) > 1 && os.Args[1] == "cleanup" {
		gc := services.GarbageCollector{
			Storage: storageBackend,
		}
		dryrun := false
		if len(os.Args) > 2 && os.Args[2] == "dryrun" {
			dryrun = true
		}
		assets, err := gc.CleanupAssets(dryrun)
		if err != nil {
			panic(err)
		}
		log.Println("Found", len(assets), "assets to cleanup")
		return
	}

	r := router.SetupGinServer()
	// Setup Cors if we are in Debug mode, otherwise UI would be under the same domain name
	if gin.Mode() == gin.DebugMode {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
		corsConfig.AllowCredentials = true
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
		r.Use(cors.New(corsConfig))
	}

	router.PackageRouter(r, storageBackend)

	err := r.Run(config.Get().ListenAddress)
	if err != nil {
		panic(err)
	}
}
