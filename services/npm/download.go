package npm

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
)

func (s *Service) DownloadPackage(c *gin.Context) {
	filename := c.Param("filename")
	pkgName, version := s.PkgVersionFromFilename(filename)
	pkg := models.Package{}
	versionInfo := models.PackageVersion{}
	db.DB().Find(&pkg, "name = ?", pkgName)
	db.DB().Find(&versionInfo, "package_id = ? AND version = ?", pkg.Id, version)

	fileData, err := s.storage.GetFile(s.PackageFilename(versionInfo.Digest))
	if err != nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	defer fileData.Close()

	c.DataFromReader(200, int64(versionInfo.Size), "application/octet-stream", fileData, map[string]string{
		"Content-Disposition": "attachment; filename=" + filename,
	})
}
