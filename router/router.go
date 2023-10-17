package router

import (
	"github.com/alin-io/pkgstore/middlewares"
	"github.com/alin-io/pkgstore/models"
	"github.com/alin-io/pkgstore/services"
	"github.com/alin-io/pkgstore/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
)

func SetupGinServer() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middlewares.AuthMiddleware)

	r.GET("/", services.HealthCheckHandler)

	r.RedirectTrailingSlash = false

	return r
}

func PackageRouter(r *gin.Engine, storageBackend storage.BaseStorageBackend) {
	initNpmRoutes(r, storageBackend)
	initPypiRoutes(r, storageBackend)
	initContainerRoutes(r, storageBackend)
	initApiRoutes(r, storageBackend)
}

func PkgNameAccessHandler(service services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := middlewares.GetAuthCtx(c)
		pkgName := service.ConstructFullPkgName(c)
		filename := c.Param("filename")
		if len(filename) > 0 && len(pkgName) == 0 {
			pkgName, _ = service.PkgVersionFromFilename(filename)
		}

		if len(authCtx.AuthId) > 0 && strings.HasPrefix(pkgName, authCtx.Namespace) {
			c.Next()
			return
		}

		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			pkg := models.Package[any]{}
			err := pkg.FillByName(pkgName, service.GetPrefix())
			if err != nil {
				service.SetAuthHeaderAndAbort(c)
				return
			}
			if pkg.ID == uuid.Nil || !pkg.IsPublic {
				service.SetAuthHeaderAndAbort(c)
				return
			}
		} else {
			service.SetAuthHeaderAndAbort(c)
			return
		}
	}
}
