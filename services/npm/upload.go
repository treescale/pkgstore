package npm

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
	"gorm.io/datatypes"
)

type npmUploadRequestBody struct {
	Attachments map[string]struct {
		ContentType string `json:"content_type"`
		Data        string `json:"data"`
		Length      int    `json:"length"`
	} `json:"_attachments"`
	Id          string                     `json:"_id"`
	Description string                     `json:"description"`
	Name        string                     `json:"name"`
	Readme      string                     `json:"readme"`
	DistTags    map[string]string          `json:"dist-tags"`
	Versions    map[string]PackageMetadata `json:"versions"`
}

func (s *Service) UploadHandler(c *gin.Context) {
	requestBody := npmUploadRequestBody{}
	authCtx := middlewares.GetAuthCtx(c)
	err := c.ShouldBind(&requestBody)
	if err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}

	decodedBytes := make([]byte, 0)
	for _, attachment := range requestBody.Attachments {
		decodedBytes, err = base64.StdEncoding.DecodeString(attachment.Data)
		if err != nil {
			c.JSON(500, gin.H{"error": "Unable to Upload Package"})
			return
		}
		break
	}

	checksum, _, err := s.ChecksumReader(bytes.NewReader(decodedBytes))
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}

	currentVersion := ""
	var pkgVersion models.PackageVersion[PackageMetadata]

	pkg := models.Package[PackageMetadata]{
		Namespace: authCtx.Namespace,
		Service:   s.Prefix,
	}
	err = pkg.FillByName(requestBody.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to check the DB for package"})
		return
	}

	if pkg.ID != uuid.Nil {
		pkgVersion, err = pkg.Version(currentVersion)
		if err != nil {
			c.JSON(500, gin.H{"error": "Unable to check the DB for package version"})
			return
		}
	} else {
		pkg = models.Package[PackageMetadata]{
			Name:      requestBody.Name,
			Service:   s.Prefix,
			AuthId:    authCtx.AuthId,
			Namespace: authCtx.Namespace,
		}
	}

	if pkgVersion.ID != uuid.Nil && len(pkgVersion.Digest) > 0 {
		if pkgVersion.Digest == checksum {
			c.JSON(200, MetadataResponse{
				Name:     pkg.Name,
				DistTags: requestBody.DistTags,
				Versions: map[string]PackageMetadata{
					pkgVersion.Version: pkgVersion.Metadata.Data(),
				},
			})
			return
		} else {
			c.JSON(400, gin.H{"error": "Wrong checksum for the existing Package Version"})
			return
		}
	}

	for _, versionInfo := range requestBody.Versions {
		currentVersion = versionInfo.Version

		pkgVersion = models.PackageVersion[PackageMetadata]{
			Version:   currentVersion,
			Digest:    checksum,
			Service:   s.Prefix,
			AuthId:    authCtx.AuthId,
			Namespace: authCtx.Namespace,
			Metadata:  datatypes.NewJSONType[PackageMetadata](versionInfo),
		}

		for tagName, tagVersion := range requestBody.DistTags {
			if tagVersion == versionInfo.Version {
				pkgVersion.Tag = tagName
				break
			}
		}
		break
	}

	err = s.Storage.WriteFile(s.PackageFilename(checksum), nil, bytes.NewReader(decodedBytes))
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}

	asset := models.Asset{
		Service: s.Prefix,
	}

	err = asset.FillByDigest(checksum)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}

	if asset.ID == uuid.Nil {
		asset = models.Asset{
			Size:        int64(len(decodedBytes)),
			Service:     s.Prefix,
			Digest:      checksum,
			UploadUUID:  uuid.NewString(),
			UploadRange: fmt.Sprintf("0-%d", len(decodedBytes)),
		}
		err = asset.Insert()
		if err != nil {
			c.JSON(500, gin.H{"error": "Unable to Upload Package"})
			return
		}
	}

	pkgVersion.Size = asset.Size
	pkgVersion.AssetIds = asset.ID.String()

	pkg.LatestVersion = pkgVersion.Version

	if pkg.ID == uuid.Nil {
		pkg.Versions = []models.PackageVersion[PackageMetadata]{pkgVersion}
		err = pkg.Insert()
	} else if len(pkgVersion.Digest) == 0 {
		err = pkg.InsertVersion(pkgVersion)
	}

	if err != nil {
		log.Println("Unable to create package in DB: ", err)
		err = s.Storage.DeleteFile(s.PackageFilename(checksum))
		if err != nil {
			log.Println("Unable to delete package from storage: ", err)
		}
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}
	c.JSON(200, MetadataResponse{
		Name:     pkg.Name,
		DistTags: requestBody.DistTags,
		Versions: map[string]PackageMetadata{
			pkgVersion.Version: pkgVersion.Metadata.Data(),
		},
	})
}
