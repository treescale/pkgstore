package router

import (
	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/services/api"
	"github.com/treescale/pkgstore/storage"
)

func initApiRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	apiService := api.NewApiService(storageBackend)
	apiRoutes := r.Group("/api")
	{
		apiRoutes.Use(middlewares.PkgNameAccessHandler(apiService))

		apiRoutes.GET("/stats", apiService.RegistryStats)
		apiRoutes.GET("/packages", apiService.ListPackagesHandler)
		apiRoutes.GET("/packages/:id", apiService.GetPackage)
		apiRoutes.GET("/packages/:id/versions", apiService.ListVersionsHandler)

		apiRoutes.DELETE("/packages/:id", apiService.DeletePackage)
		apiRoutes.DELETE("/packages/:id/versions/:versionId", apiService.DeleteVersion)
	}
}
