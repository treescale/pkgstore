package router

import (
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/services/npm"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NPMRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	npmService := npm.NewService(storageBackend)
	r.GET("/*path", HandleFetch(npmService))
	r.PUT("/*path", HandleUpload(npmService))
}

func HandleUpload(routeServices ...services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, service := range routeServices {
			if service.ShouldHandleRequest(c) {
				service.UploadHandler(c)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
	}
}

func HandleFetch(routeServices ...services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, service := range routeServices {
			if service.ShouldHandleRequest(c) {
				packageName, fileName := service.PkgInfoFromRequestPath(c)
				if len(fileName) > 0 && len(packageName) > 0 {
					service.DownloadHandler(c)
				} else {
					service.MetadataHandler(c)
				}
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
	}
}
