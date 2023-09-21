package router

import (
	"github.com/alin-io/pkgproxy/services/npm"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
)

func NPMRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	npmService := npm.NewService(storageBackend)
	npmRoutes := r.Group("/:pkgName")
	{
		npmRoutes.GET("/-/:filename", npmService.DownloadPackage)
		npmRoutes.GET("", npmService.FetchMetadata)
		npmRoutes.PUT("", npmService.UploadHandler)
	}
}
