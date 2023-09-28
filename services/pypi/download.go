package pypi

import (
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

func (s *Service) DownloadHandler(c *gin.Context) {
	filename := c.GetString("filename")
	pkgName, version := s.PkgVersionFromFilename(filename)
	pkg := models.Package[pypiPackageMetadata]{}
	versionInfo := models.PackageVersion[pypiPackageMetadata]{}
	err := pkg.FillByName(pkgName, s.Prefix)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}

	versionInfo, err = pkg.Version(version)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}

	fileData, err := s.Storage.GetFile(s.PackageFilename(versionInfo.Digest, s.FilenamePostfix(filename, pkgName, version)))
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
