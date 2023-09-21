package pypi

import (
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
)

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{Prefix: "pypi", Storage: storage},
	}
}

func (s *Service) ShouldHandleRequest(c *gin.Context) bool {
	pkgPath := c.Param("path")
	return len(pkgPath) == 0 && len(c.PostForm("name")) > 0 && len(c.PostForm("version")) > 0
}
