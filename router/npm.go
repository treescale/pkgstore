package router

import (
	"fmt"
	"github.com/alin-io/pkgstore/config"
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

			npmRoutes.GET(pkgNameParam, npmService.MetadataHandler)
			npmRoutes.GET(pkgNameParam+"/-/:filename", npmService.DownloadHandler)

			npmRoutes.PUT(pkgNameParam, npmService.UploadHandler)
		}
	}
}
