package npm

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
)

type MetadataResponse struct {
	Name     string                                   `json:"name"`
	DistTags map[string]string                        `json:"dist-tags"`
	Versions map[string]models.PackageVersionMetadata `json:"versions"`
}

func (s *Service) FetchMetadata(c *gin.Context) {
	pkgName := c.Param("pkgName")
	pkg := models.Package{}
	versions := make([]models.PackageVersion, 0)
	db.DB().Find(&pkg, "name = ?", pkgName)
	db.DB().Find(&versions, "package_id = ?", pkg.Id)
	result := MetadataResponse{Name: pkgName, DistTags: make(map[string]string), Versions: make(map[string]models.PackageVersionMetadata)}

	for _, version := range versions {
		result.Versions[version.Version] = version.Metadata.Data()
		if len(version.Tag) > 0 {
			result.DistTags[version.Tag] = version.Version
		}
	}
	c.JSON(200, result)
}
