package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/config"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/services/pypi"
	"github.com/treescale/pkgstore/storage"
)

func initPypiRoutes(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	pypiService := pypi.NewService(storageBackend)
	pypiRoutes := r.Group("/pypi")
	{
		pypiRoutes.Use(middlewares.PkgNameAccessHandler(pypiService))

		pkgNameParam := ""
		for i := 0; i < config.NumberOfPkgNameLevels; i++ {
			pkgNameParam += fmt.Sprintf("/:name%d", i)
			pypiRoutes.GET(
				"/simple/"+pkgNameParam,
				pypiService.MetadataHandler,
			)
		}

		pypiRoutes.GET("/files/:sha256/:filename", pypiService.DownloadHandler)

		pypiRoutes.POST("", pypiService.UploadHandler)
	}
}
