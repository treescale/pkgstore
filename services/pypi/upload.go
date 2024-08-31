package pypi

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
	"gorm.io/datatypes"
)

func (s *Service) UploadHandler(c *gin.Context) {
	pkgName := c.PostForm("name")
	pkgVersionName := c.PostForm("version")
	authCtx := middlewares.GetAuthCtx(c)
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
	storageFilename := s.PackageFilename(checksum)

	packageModel := models.Package[PackageMetadata]{
		Namespace: authCtx.Namespace,
		Service:   s.Prefix,
	}
	pkgVersion := models.PackageVersion[PackageMetadata]{
		Namespace: authCtx.Namespace,
		Service:   s.Prefix,
	}
	_ = packageModel.FillByName(pkgName)
	if packageModel.ID != uuid.Nil {
		pkgVersion, err = packageModel.Version(pkgVersionName)
		if err != nil {
			log.Println("Unable to fill package versions: ", err)
			c.JSON(500, gin.H{"error": "Unable to Upload Package"})
			return
		}

		if pkgVersion.ID == uuid.Nil {
			pkgVersion = models.PackageVersion[PackageMetadata]{
				PackageId: packageModel.ID,
				Service:   s.Prefix,
				Digest:    checksum,
				Version:   pkgVersionName,
				Tag:       pkgVersionName,
				AuthId:    authCtx.AuthId,
				Namespace: authCtx.Namespace,
				Metadata: datatypes.NewJSONType(PackageMetadata{
					RequiresPython: c.PostForm("requires_python"),
					OriginalFiles:  []string{file.Filename},
				}),
			}

			err = pkgVersion.Save()
			if err != nil {
				log.Println("Unable to Save package versions: ", err)
				c.JSON(500, gin.H{"error": "Unable to Upload Package"})
				return
			}

			packageModel.LatestVersion = pkgVersion.Version
			_ = packageModel.Save()
		}
	}

	err = s.Storage.WriteFile(storageFilename, nil, fileHandle)
	if err != nil {
		log.Println("Unable to write package to storage: ", err)
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}

	asset := models.Asset{
		Service: s.Prefix,
	}

	err = asset.FillByDigest(checksum)
	if err != nil {
		log.Println("Unable to fill asset by digest: ", err)
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}

	if asset.ID == uuid.Nil {
		asset = models.Asset{
			Size:        size,
			Service:     s.Prefix,
			Digest:      checksum,
			UploadUUID:  uuid.NewString(),
			UploadRange: fmt.Sprintf("0-%d", size),
		}

		err = asset.Insert()
		if err != nil {
			log.Println("Unable to insert asset: ", err)
			c.JSON(500, gin.H{"error": "Unable to Upload Package"})
			return
		}
	}

	if packageModel.ID != uuid.Nil && len(pkgVersion.Digest) > 0 {
		if slices.Contains(pkgVersion.Metadata.Data().OriginalFiles, file.Filename) {
			c.JSON(200, pkgVersion)
			return
		} else {
			versionMeta := pkgVersion.Metadata.Data()
			versionMeta.OriginalFiles = append(versionMeta.OriginalFiles, file.Filename)
			pkgVersion.Metadata = datatypes.NewJSONType(versionMeta)
			pkgVersion.Size += asset.Size
			err = pkgVersion.Save()
			if err != nil {
				log.Println("Unable to update package version metadata: ", err)
				c.JSON(500, gin.H{"error": "Unable to Upload Package"})
				return
			}
			err = pkgVersion.AddAsset(&asset)
			if err != nil {
				log.Println("Unable to add asset to package version: ", err)
				c.JSON(500, gin.H{"error": "Unable to Upload Package"})
				return
			}
		}
	} else {
		pkgVersion = models.PackageVersion[PackageMetadata]{
			Service:   s.Prefix,
			Digest:    checksum,
			Version:   pkgVersionName,
			AuthId:    authCtx.AuthId,
			Namespace: authCtx.Namespace,
			Size:      asset.Size,
			Metadata: datatypes.NewJSONType(PackageMetadata{
				RequiresPython: c.PostForm("requires_python"),
				OriginalFiles:  []string{file.Filename},
			}),
			AssetIds: asset.ID.String(),
		}

		packageModel = models.Package[PackageMetadata]{
			Name:          pkgName,
			Service:       s.Prefix,
			AuthId:        authCtx.AuthId,
			Namespace:     authCtx.Namespace,
			LatestVersion: pkgVersionName,
			Versions: []models.PackageVersion[PackageMetadata]{
				pkgVersion,
			},
		}

		err = packageModel.Insert()
		if err != nil {
			log.Println("Unable to create package in DB: ", err)
		}
	}

	if err != nil {
		err = s.Storage.DeleteFile(storageFilename)
		if err != nil {
			log.Println("Unable to Delete/Rollback package upload: ", err)
		}
		c.JSON(500, gin.H{"error": "Unable to Upload Package"})
		return
	}
	c.JSON(200, pkgVersion.Metadata.Data())
}
