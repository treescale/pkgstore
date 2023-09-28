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

func SetupGinServer() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	return r
}

func PackageRouter(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	r.Use(middlewares.AuthMiddleware)

	npmService := npm.NewService(storageBackend)
	pypiService := pypi.NewService(storageBackend)

	r.GET("/", services.HealthCheckHandler)

	r.GET("npm/*path", HandleFetch(npmService))
	r.GET("pypi/*path", HandleFetch(pypiService))

	r.PUT("npm/*path", npmService.UploadHandler)
	r.POST("pypi/*path", pypiService.UploadHandler)
}

func HandleFetch(service services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		packageName, fileName := service.PkgInfoFromRequestPath(c)
		c.Set("pkgName", packageName)
		c.Set("filename", fileName)

		if len(fileName) > 0 && len(packageName) > 0 {
			service.DownloadHandler(c)
		} else {
			service.MetadataHandler(c)
		}
		c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
	}
}
