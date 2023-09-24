package router

import (
	"github.com/alin-io/pkgproxy/middlewares"
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/services/npm"
	"github.com/alin-io/pkgproxy/services/pypi"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupGinServer(storageBackend storage.BaseStorageBackend) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	PackageRouter(r, storageBackend)

	return r
}

func PackageRouter(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	r.Use(middlewares.AuthMiddleware)

	npmService := npm.NewService(storageBackend)
	pypiService := pypi.NewService(storageBackend)
	r.GET("/*path", HandleFetch(npmService, pypiService))
	r.PUT("/*path", HandleUpload(npmService, pypiService))
	r.POST("/*path", HandleUpload(npmService, pypiService))
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
				c.Set("pkgName", packageName)
				c.Set("filename", fileName)

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
