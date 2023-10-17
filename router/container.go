package router

import (
	"fmt"
	"github.com/alin-io/pkgstore/config"
	"github.com/alin-io/pkgstore/services/container"
	"github.com/alin-io/pkgstore/storage"
	"github.com/gin-gonic/gin"
)

func initContainerRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	containerService := container.NewService(storageBackend)
	containerRoutes := r.Group("/v2")
	{
		containerRoutes.GET("/", func(context *gin.Context) {
			context.JSON(200, gin.H{"status": "ok"})
		})

		pkgNameParam := ""

		for i := 0; i < config.NumberOfPkgNameLevels; i++ {
			pkgNameParam += fmt.Sprintf("/:name%d", i)

			// Upload Process
			containerRoutes.GET(pkgNameParam+"/blobs/uploads/:uuid", containerService.GetUploadProgressHandler)
			containerRoutes.HEAD(pkgNameParam+"/blobs/:sha256", containerService.CheckBlobExistenceHandler)
			containerRoutes.POST(pkgNameParam+"/blobs/uploads/", containerService.StartLayerUploadHandler)
			containerRoutes.PATCH(pkgNameParam+"/blobs/uploads/:uuid", containerService.ChunkUploadHandler)
			containerRoutes.PUT(pkgNameParam+"/blobs/uploads/:uuid", containerService.UploadHandler)
			containerRoutes.PUT(pkgNameParam+"/manifests/:reference", containerService.ManifestUploadHandler)

			// Download Process
			containerRoutes.GET(pkgNameParam+"/manifests/:reference", containerService.MetadataHandler)
			containerRoutes.HEAD(pkgNameParam+"/manifests/:reference", containerService.CheckMetadataHandler)
			containerRoutes.GET(pkgNameParam+"/blobs/:sha256", containerService.DownloadHandler)
		}
	}
}
