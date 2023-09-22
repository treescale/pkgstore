package npm

import (
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
	"regexp"
)

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{Prefix: "npm", Storage: storage},
	}
}

func (s *Service) PkgInfoFromRequestPath(c *gin.Context) (pkgName string, filename string) {
	pkgPath := c.Param("path")

	// /:pkgName/-/:filename
	// /@orgname/pkgName/-/:filename
	pattern := `^/(?P<pkgName>(@[^/]+/)?[^/]+)(?:/-/)(?P<filename>[^/]+)$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(pkgPath)
	if matches == nil {
		return "", ""
	}

	for i, name := range re.SubexpNames() {
		if name == "pkgName" {
			pkgName = matches[i]
		} else if name == "filename" {
			filename = matches[i]
		}
	}

	return pkgName, filename
}
