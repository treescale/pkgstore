package npm

import (
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
)

func (s *Service) DownloadHandler(c *gin.Context) {
	filename := c.Param("filename")
	pkgName, _ := s.ConstructFullPkgName(c)
	namespace := middlewares.GetAuthCtx(c).Namespace

	_, version := s.PkgVersionFromFilename(filename)
	pkg := models.Package[PackageMetadata]{
		Namespace: namespace,
		Service:   s.Prefix,
	}
	versionInfo := models.PackageVersion[PackageMetadata]{
		Namespace: namespace,
		Service:   s.Prefix,
	}
	err := pkg.FillByName(pkgName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}

	if pkg.ID == uuid.Nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	versionInfo, err = pkg.Version(version)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}

	if len(versionInfo.Digest) == 0 {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	fileAssets, err := versionInfo.GetAssets()
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}

	if len(fileAssets) == 0 {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	fileAsset := fileAssets[0]

	if fileAsset.ID == uuid.Nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	fileData, err := s.Storage.GetFile(s.PackageFilename(fileAsset.Digest))
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

	c.DataFromReader(200, int64(fileAsset.Size), "application/octet-stream", fileData, map[string]string{
		"Content-Disposition": "attachment; filename=" + filename,
	})
}
