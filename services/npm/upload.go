package npm

import (
	"bytes"
	"encoding/base64"
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type npmUploadRequestBody struct {
	Attachments map[string]struct {
		ContentType string `json:"content_type"`
		Data        string `json:"data"`
		Length      int    `json:"length"`
	} `json:"_attachments"`
	Id          string                        `json:"_id"`
	Description string                        `json:"description"`
	Name        string                        `json:"name"`
	Readme      string                        `json:"readme"`
	DistTags    map[string]string             `json:"dist-tags"`
	Versions    map[string]npmPackageMetadata `json:"versions"`
}

func (s *Service) UploadHandler(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	requestBody := npmUploadRequestBody{}
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
	var pkgVersion models.PackageVersion[npmPackageMetadata]
	for _, versionInfo := range requestBody.Versions {
		currentVersion = versionInfo.Version

		pkgVersion = models.PackageVersion[npmPackageMetadata]{
			Version:  currentVersion,
			Digest:   checksum,
			Metadata: datatypes.NewJSONType[npmPackageMetadata](versionInfo),
			Size:     uint64(len(decodedBytes)),
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

	db.DB().Create(&models.Package[npmPackageMetadata]{
		Name:      requestBody.Name,
		Service:   s.Prefix,
		Namespace: "",
		AuthId:    c.GetString("token"),
		Versions:  []models.PackageVersion[npmPackageMetadata]{pkgVersion},
	})
	c.JSON(200, requestBody)
}
