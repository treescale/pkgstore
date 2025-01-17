package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/config"
	"github.com/treescale/pkgstore/models"
	"github.com/treescale/pkgstore/services"
)

func PkgNameAccessHandler(service services.PackageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pkgName, namespace := service.ConstructFullPkgName(c)
		filename := c.Param("filename")
		if len(filename) > 0 && len(pkgName) == 0 {
			pkgName, _ = service.PkgVersionFromFilename(filename)
		}

		pkgAction := "pull"
		if c.Request.Method == "PUT" || c.Request.Method == "POST" || c.Request.Method == "PATCH" {
			pkgAction = "push"
		}

		authResult := &AuthResult{}

		if len(config.Get().AuthEndpoint) > 0 {
			tokenString, err := extractTokenHeader(c)
			if err != nil {
				service.SetAuthHeaderAndAbort(c, "Unable to get the token")
				return
			}

			authResult, err = getRemoteAuthContext(c, pkgName, tokenString, service.GetPrefix(), pkgAction)
			if err != nil {
				service.SetAuthHeaderAndAbort(c, authResult.Error)
				return
			}

			if len(authResult.Namespace) == 0 {
				service.AbortRequestWithError(c, 403, "Unable to get the namespace from the auth endpoint")
				return
			}

			if pkgAction == "push" && authResult.Namespace != namespace {
				service.AbortRequestWithError(c, 403, "You don't have access to this namespace")
				return
			}

			if len(pkgName) > 0 {
				pkg := models.Package[any]{
					Namespace: authResult.Namespace,
					Service:   service.GetPrefix(),
				}
				err := pkg.FillByName(pkgName)
				if err != nil {
					service.AbortRequestWithError(c, 500, "Unable to check the DB for the package")
					return
				}

				if pkg.ID != uuid.Nil {
					if pkgAction == "pull" {
						authResult.PublicAccess = pkg.IsPublic
					}

					if pkg.Namespace != authResult.Namespace && !pkg.IsPublic {
						service.AbortRequestWithError(c, 401, "You don't have access to this package")
						return
					}
				}
			}
		} else {
			authResult.PublicAccess = true
			authResult.AuthId = AuthIdPublic
		}

		if len(authResult.AuthId) == 0 {
			service.AbortRequestWithError(c, 401, "Unauthorized")
			return
		}

		if !authResult.PublicAccess && !strings.HasPrefix(pkgName, authResult.Namespace) {
			service.AbortRequestWithError(c, 401, "You don't have access to this package")
			return
		}

		c.Set("auth", authResult)
		c.Next()
	}
}
