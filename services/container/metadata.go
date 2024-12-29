package container

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
)

func (s *Service) MetadataHandler(c *gin.Context) {
	_, pkgVersion := s.pkgVersionMetadata(c)
	if pkgVersion.ID == uuid.Nil {
		return
	}

	metadata := pkgVersion.Metadata.Data()

	c.Header("Docker-Content-Digest", "sha256:"+pkgVersion.Digest)
	c.Data(200, metadata.ContentType, metadata.MetadataBuffer)
}

func (s *Service) CheckMetadataHandler(c *gin.Context) {
	_, pkgVersion := s.pkgVersionMetadata(c)
	if pkgVersion.ID == uuid.Nil {
		return
	}

	metadata := pkgVersion.Metadata.Data()

	c.Header("Content-Type", metadata.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", len(metadata.MetadataBuffer)))
	c.Header("Docker-Content-Digest", fmt.Sprintf("sha256:%s", pkgVersion.Digest))
	c.Status(200)
	c.Done()
}

func (s *Service) pkgVersionMetadata(c *gin.Context) (pkg models.Package[PackageMetadata], pkgVersion models.PackageVersion[PackageMetadata]) {
	namespace := middlewares.GetAuthCtx(c).Namespace
	name, _ := s.ConstructFullPkgName(c)
	tagOrDigest := c.Param("reference")
	pkg = models.Package[PackageMetadata]{
		Namespace: namespace,
		Service:   s.Prefix,
	}
	err := pkg.FillByName(name)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Error while trying to get package info",
				},
			},
		})
		return
	}
	if pkg.ID == uuid.Nil {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Package not found",
				},
			},
		})
		return
	}
	if strings.Contains(tagOrDigest, "sha256:") {
		pkgVersion.Namespace = namespace
		pkgVersion.Service = s.Prefix
		err = pkgVersion.FillByDigest(strings.Replace(tagOrDigest, "sha256:", "", 1))
	} else {
		pkgVersion, err = pkg.Version(tagOrDigest)
	}
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Error while trying to get package info",
				},
			},
		})
		return
	}
	if pkgVersion.ID == uuid.Nil || pkgVersion.PackageId != pkg.ID {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Package version not found",
				},
			},
		})
		return
	}
	return
}
