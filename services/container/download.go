package container

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
)

func (s *Service) DownloadHandler(c *gin.Context) {
	name, _ := s.ConstructFullPkgName(c)
	inputDigest := c.Param("sha256")
	authCtx := middlewares.GetAuthCtx(c)
	digest := strings.Replace(inputDigest, "sha256:", "", 1)
	pkg := models.Package[PackageMetadata]{
		Namespace: authCtx.Namespace,
		Service:   s.Prefix,
	}
	err := pkg.FillByName(name)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error while trying to get package info"})
		return
	}
	if pkg.ID == uuid.Nil {
		c.JSON(404, gin.H{"error": "Package not found"})
		return
	}

	asset := models.Asset{
		Service: s.Prefix,
	}
	err = asset.FillByDigest(digest)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to check the DB for package version"})
		return
	}
	if asset.Digest != digest {
		c.JSON(404, gin.H{"error": "Uploaded asset not found"})
		return
	}

	fileData, err := s.Storage.GetFile(s.PackageFilename(asset.Digest))
	if err != nil || fileData == nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}

	defer fileData.Close()

	// Set headers
	c.Header("Content-Disposition", "attachment; filename="+digest)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", asset.Size))

	// Stream the file data
	_, err = io.Copy(c.Writer, fileData)
	if err != nil {
		log.Printf("Error streaming file: %v", err)
		c.AbortWithStatus(500)
		return
	}
}
