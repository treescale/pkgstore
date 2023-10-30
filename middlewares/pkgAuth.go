package middlewares

import (
	"github.com/alin-io/pkgstore/config"
	"github.com/alin-io/pkgstore/models"
	"github.com/alin-io/pkgstore/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func PkgNameAccessHandler(service services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pkgName := service.ConstructFullPkgName(c)
		filename := c.Param("filename")
		if len(filename) > 0 && len(pkgName) == 0 {
			pkgName, _ = service.PkgVersionFromFilename(filename)
		}

		pkgAction := "pull"
		if c.Request.Method == "PUT" || c.Request.Method == "POST" || c.Request.Method == "PATCH" {
			pkgAction = "push"
		}

		authResult := &AuthResult{}

		if pkgAction == "pull" {
			pkg := models.Package[any]{}
			err := pkg.FillByName(pkgName, service.GetPrefix())
			if err != nil {
				service.AbortRequestWithError(c, 500, "Unable to check the DB for package version")
				return
			}
			if pkg.ID != uuid.Nil && pkg.IsPublic {
				authResult = &AuthResult{
					PublicAccess: true,
					AuthId:       AuthIdPublic,
				}
			}
		} else if len(config.Get().AuthEndpoint) > 0 {
			tokenString, err := extractTokenHeader(c)
			if err != nil {
				service.AbortRequestWithError(c, 401, err.Error())
			}

			authResult, err = getRemoteAuthContext(c, pkgName, tokenString, service.GetPrefix(), pkgAction)
			if err != nil {
				service.AbortRequestWithError(c, 401, err.Error())
			}
		}

		if len(authResult.AuthId) == 0 {
			service.AbortRequestWithError(c, 401, "Unauthorized")
		}

		c.Set("auth", authResult)
		c.Next()
	}
}
