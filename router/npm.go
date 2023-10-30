package router

import (
	"fmt"
	"github.com/alin-io/pkgstore/config"
	"github.com/alin-io/pkgstore/middlewares"
	"github.com/alin-io/pkgstore/services/npm"
	"github.com/alin-io/pkgstore/storage"
	"github.com/gin-gonic/gin"
)

func initNpmRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	npmService := npm.NewService(storageBackend)

	npmRoutes := r.Group("/npm")
	{
		pkgNameParam := ""
		for i := 0; i < config.NumberOfPkgNameLevels; i++ {
			pkgNameParam += fmt.Sprintf("/:name%d", i)

			pkgNameRoutes := npmRoutes.Group(pkgNameParam)
			{
				pkgNameRoutes.Use(middlewares.PkgNameAccessHandler(npmService))

				pkgNameRoutes.GET("", npmService.MetadataHandler)
				pkgNameRoutes.GET("-/:filename", npmService.DownloadHandler)

				pkgNameRoutes.PUT("", npmService.UploadHandler)
			}
		}
	}
}
