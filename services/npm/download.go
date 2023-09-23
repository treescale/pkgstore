package npm

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

func (s *Service) DownloadHandler(c *gin.Context) {
	filename := c.Param("filename")
	pkgName, version := s.PkgVersionFromFilename(filename)
	pkg := models.Package[npmPackageMetadata]{}
	versionInfo := models.PackageVersion[npmPackageMetadata]{}
	db.DB().Find(&pkg, "name = ?", pkgName)
	db.DB().Find(&versionInfo, "package_id = ? AND version = ?", pkg.Id, version)

	fileData, err := s.Storage.GetFile(s.PackageFilename(versionInfo.Digest))
	if err != nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	defer func(fileData io.ReadCloser) {
		err := fileData.Close()
		if err != nil {
			log.Println(err)
		}
	}(fileData)

	c.DataFromReader(200, int64(versionInfo.Size), "application/octet-stream", fileData, map[string]string{
		"Content-Disposition": "attachment; filename=" + filename,
	})
}
