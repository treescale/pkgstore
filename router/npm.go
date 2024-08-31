package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/config"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/services/npm"
	"github.com/treescale/pkgstore/storage"
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
