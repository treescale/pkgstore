package pypi

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"io"
	"log"
	"mime/multipart"
	"strings"
)

func (s *Service) UploadHandler(c *gin.Context) {
	pkgName := c.PostForm("name")
	pkgVersionName := c.PostForm("version")
	file, err := c.FormFile("content")
	if err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	fileHandle, err := file.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}

	defer func(fileHandle multipart.File) {
		err := fileHandle.Close()
		if err != nil {
			log.Println(err)
		}
	}(fileHandle)

	checksum, size, err := s.ChecksumReader(fileHandle)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}
	_, err = fileHandle.Seek(0, io.SeekStart)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}
	err = s.Storage.WriteFile(s.PackageFilename(checksum), nil, fileHandle)

	pkgVersion := models.PackageVersion{
		Digest:  checksum,
		Version: pkgVersionName,
		Size:    uint64(size),
		Metadata: datatypes.NewJSONType(models.PackageVersionMetadata{
			RequiresPython:  c.PostForm("requires_python"),
			FilenamePostfix: strings.Replace(file.Filename, s.constructPackageOriginalFilename(pkgName, pkgVersionName, ""), "", 1),
		}),
	}

	db.DB().Create(&models.Package{
		Name:      pkgName,
		Service:   s.Prefix,
		Namespace: "",
		AuthId:    c.GetString("token"),
		Versions: []models.PackageVersion{
			pkgVersion,
		},
	})
	c.JSON(200, pkgVersion)
}
