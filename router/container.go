package router

import (
	"fmt"
	"strings"

	"github.com/alin-io/pkgstore/config"
	"github.com/alin-io/pkgstore/middlewares"
	"github.com/alin-io/pkgstore/services/container"
	"github.com/alin-io/pkgstore/storage"
	"github.com/gin-gonic/gin"
)

func initContainerRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	containerService := container.NewService(storageBackend)
	containerRoutes := r.Group("/v2")
	{
		containerRoutes.GET("/", func(c *gin.Context) {
			authToken := c.GetHeader("Authorization")
			if len(authToken) > 0 {
				c.JSON(200, gin.H{"token": strings.Split(authToken, " ")[1]})
			} else {
				containerService.SetAuthHeaderAndAbort(c)
			}
		})

		pkgNameParam := ""

		for i := 0; i < config.NumberOfPkgNameLevels; i++ {
			pkgNameParam += fmt.Sprintf("/:name%d", i)
			pkgNameRoutes := containerRoutes.Group(pkgNameParam)
			{
				pkgNameRoutes.Use(middlewares.PkgNameAccessHandler(containerService))

				// Upload Process
				pkgNameRoutes.GET("blobs/uploads/:uuid", containerService.GetUploadProgressHandler)
				pkgNameRoutes.HEAD("blobs/:sha256", containerService.CheckBlobExistenceHandler)
				pkgNameRoutes.POST("blobs/uploads/", containerService.StartLayerUploadHandler)
				pkgNameRoutes.PATCH("blobs/uploads/:uuid", containerService.ChunkUploadHandler)
				pkgNameRoutes.PUT("blobs/uploads/:uuid", containerService.UploadHandler)
				pkgNameRoutes.PUT("manifests/:reference", containerService.ManifestUploadHandler)

				// Download Process
				pkgNameRoutes.GET("manifests/:reference", containerService.MetadataHandler)
				pkgNameRoutes.HEAD("manifests/:reference", containerService.CheckMetadataHandler)
				pkgNameRoutes.GET("blobs/:sha256", containerService.DownloadHandler)
			}
		}
	}
}
