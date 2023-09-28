package npm

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
)

type MetadataResponse struct {
	Name     string                        `json:"name"`
	DistTags map[string]string             `json:"dist-tags"`
	Versions map[string]npmPackageMetadata `json:"versions"`
}

func (s *Service) MetadataHandler(c *gin.Context) {
	pkgName := c.GetString("pkgName")
	pkg := models.Package[npmPackageMetadata]{}
	versions := make([]models.PackageVersion[npmPackageMetadata], 0)
	db.DB().Find(&pkg, "name = ?", pkgName)
	db.DB().Find(&versions, "package_id = ?", pkg.Id)
	if pkg.Id < 1 || len(versions) == 0 {
		s.ProxyToPublicRegistry(c)
		return
	}
	result := MetadataResponse{Name: pkgName, DistTags: make(map[string]string), Versions: make(map[string]npmPackageMetadata)}

	for _, version := range versions {
		result.Versions[version.Version] = version.Metadata.Data()
		if len(version.Tag) > 0 {
			result.DistTags[version.Tag] = version.Version
		}
	}
	c.JSON(200, result)
}
