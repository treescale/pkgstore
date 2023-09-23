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
	filenamePostfix := strings.Replace(file.Filename, s.constructPackageOriginalFilename(pkgName, pkgVersionName, ""), "", 1)
	storageFilename := s.PypiPackageFilename(checksum, filenamePostfix)
	err = s.Storage.WriteFile(storageFilename, nil, fileHandle)

	pkgVersion := models.PackageVersion[pypiPackageMetadata]{
		Digest:  checksum,
		Version: pkgVersionName,
		Size:    uint64(size),
		Metadata: datatypes.NewJSONType(pypiPackageMetadata{
			RequiresPython:  c.PostForm("requires_python"),
			FilenamePostfix: filenamePostfix,
		}),
	}

	err = db.DB().Create(&models.Package[pypiPackageMetadata]{
		Name:      pkgName,
		Service:   s.Prefix,
		Namespace: "",
		AuthId:    c.GetString("token"),
		Versions: []models.PackageVersion[pypiPackageMetadata]{
			pkgVersion,
		},
	}).Error
	if err != nil {
		log.Println("Unable to create package in DB: ", err)
		err = s.Storage.DeleteFile(storageFilename)
		if err != nil {
			log.Println("Unable to Delete/Rollback package upload: ", err)
		}
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}
	c.JSON(200, pkgVersion)
}
