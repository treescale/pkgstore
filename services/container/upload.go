package container

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
	"gorm.io/datatypes"
)

// StartLayerUploadHandler POST /v2/<name>/blobs/uploads/
func (s *Service) StartLayerUploadHandler(c *gin.Context) {
	pkgName, _ := s.ConstructFullPkgName(c)
	asset := models.Asset{
		Service: s.Prefix,
	}
	err := asset.StartUpload()
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to Start the upload process",
				},
			},
		})
		return
	}

	c.Header("Location", "/v2/"+pkgName+"/blobs/uploads/"+asset.UploadUUID)
	c.Header("Docker-Upload-UUID", asset.UploadUUID)
	c.Header("Range", "bytes="+asset.UploadRange)
	c.Header("Content-Length", "0")
	c.Status(202)
	c.Done()
}

func (s *Service) CheckBlobExistenceHandler(c *gin.Context) {
	digest := strings.Replace(c.Param("sha256"), "sha256:", "", 1)
	asset := models.Asset{
		Service: s.Prefix,
	}
	err := asset.FillByDigest(digest)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to check the DB for package version",
				},
			},
		})
		return
	}

	if asset.ID == uuid.Nil {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Blob not found",
				},
			},
		})
		return
	}

	c.Header("Docker-Content-Digest", "sha256:"+digest)
	c.Header("Content-Length", fmt.Sprintf("%d", asset.Size))
	c.Status(200)
	c.Done()
}

func (s *Service) GetUploadProgressHandler(c *gin.Context) {
	pkgName, _ := s.ConstructFullPkgName(c)
	uploadUUID := c.Param("uuid")
	asset := models.Asset{
		Service: s.Prefix,
	}
	err := asset.FillByUploadUUID(uploadUUID)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to get upload progress",
				},
			},
		})
		log.Println("Unable to get upload progress", err)
		return
	}
	if asset.UploadUUID != uploadUUID {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Upload not found",
				},
			},
		})
		log.Println("Upload not found")
		return
	}

	c.Header("Range", asset.UploadRange)
	c.Header("Location", "/v2/"+pkgName+"/blobs/uploads/"+uploadUUID)
	c.Header("Docker-Upload-UUID", uploadUUID)
	c.Status(204)
	c.Done()
}

func (s *Service) ChunkUploadHandler(c *gin.Context) {
	pkgName, _ := s.ConstructFullPkgName(c)
	uploadUUID := c.Param("uuid")
	asset := models.Asset{
		Service: s.Prefix,
	}
	err := asset.FillByUploadUUID(uploadUUID)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to get upload progress",
				},
			},
		})
		log.Println("Unable to get upload progress", err)
		return
	}
	if asset.UploadUUID != uploadUUID {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Upload not found",
				},
			},
		})
		return
	}

	_, chunkSize, err := s.appendStorageData(uploadUUID, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to save chunk",
				},
			},
		})
		log.Println("Unable to save chunk metadata", err)
		return
	}

	if chunkSize > 0 {
		asset.UploadRange = fmt.Sprintf("%d-%d", asset.Size, chunkSize)
		asset.Size = chunkSize
	}
	err = asset.Update()
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to save chunk metadata",
				},
			},
		})
		log.Println("Unable to save chunk metadata", err)
		return
	}

	c.Header("Location", "/v2/"+pkgName+"/blobs/uploads/"+uploadUUID)
	c.Header("Docker-Upload-UUID", uploadUUID)
	c.Header("Range", asset.UploadRange)
	c.Header("Content-Length", "0")
	c.Status(204)
	c.Done()
}

func (s *Service) UploadHandler(c *gin.Context) {
	pkgName, _ := s.ConstructFullPkgName(c)
	inputDigest := strings.Replace(c.Query("digest"), "sha256:", "", 1)
	uploadUUID := c.Param("uuid")
	asset := models.Asset{
		Service: s.Prefix,
	}
	err := asset.FillByUploadUUID(uploadUUID)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to get upload progress",
				},
			},
		})
		return
	}
	if asset.UploadUUID != uploadUUID {
		c.JSON(404, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Upload not found",
				},
			},
		})
		return
	}

	digest, totalSize, err := s.appendStorageData(uploadUUID, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to save chunk",
				},
			},
		})
		return
	}

	if inputDigest != "" && inputDigest != digest {
		c.JSON(400, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Digest mismatch",
				},
			},
		})
		return
	}

	err = s.Storage.CopyFile(s.PackageFilename(uploadUUID), s.PackageFilename(digest))
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to store the file",
				},
			},
		})
		return
	}

	asset2 := models.Asset{
		Service: s.Prefix,
	}
	err = asset2.FillByDigest(digest)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to check the DB for package version",
				},
			},
		})
		return
	}
	if asset2.ID == uuid.Nil || asset2.Digest != digest {
		asset.Digest = digest
		asset.Size = totalSize
		err = asset.Update()
		if err != nil {
			c.JSON(500, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable to save chunk metadata",
					},
				},
			})
			return
		}
	}

	err = s.Storage.DeleteFile(s.PackageFilename(uploadUUID))
	if err != nil {
		log.Println(err)
	}

	c.Header("Location", "/v2/"+pkgName+"/blobs/"+digest)
	c.Header("Content-Range", "0-"+fmt.Sprintf("%d", totalSize))
	c.Header("Content-Length", "0")
	c.Header("Docker-Content-Digest", "sha256:"+digest)
	c.Status(204)
	c.Done()
}

func (s *Service) ManifestUploadHandler(c *gin.Context) {
	var (
		tagName    = strings.Replace(c.Param("reference"), "sha256:", "", 1)
		pkgName, _ = s.ConstructFullPkgName(c)
		err        error
	)

	authCtx := middlewares.GetAuthCtx(c)

	metadataBody, _ := io.ReadAll(c.Request.Body)

	hasher := sha256.New()
	_, _ = hasher.Write(metadataBody)
	digest := hex.EncodeToString(hasher.Sum(nil))

	metadata := PackageMetadata{
		ContentType:    c.Request.Header.Get("Content-Type"),
		MetadataBuffer: metadataBody,
		Digest:         digest,
	}

	versionSize := int64(0)
	assets := make([]models.Asset, 0)

	switch metadata.ContentType {
	case ManifestV1ContentType:
		manifest := ManifestV1{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			c.JSON(400, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable tօ parse the manifest",
					},
				},
			})
			return
		}
		tagName = manifest.Tag
		pkgName = manifest.Name
		if len(manifest.FsLayers) == 0 {
			c.JSON(400, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "No layers found",
					},
				},
			})
			return
		}
		for _, layer := range manifest.FsLayers {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(layer.BlobSum, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				c.JSON(500, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Unable to check the DB for package version",
						},
					},
				})
				return
			}
			if asset.Digest != layerDigest {
				c.JSON(404, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Uploaded asset not found",
						},
					},
				})
				return
			}
			versionSize += asset.Size
			assets = append(assets, asset)
		}
	case ManifestV2ContentType, ManifestOCIV1ContentType:
		manifest := ManifestV2{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			c.JSON(400, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable tօ parse the manifest",
					},
				},
			})
			return
		}
		for _, layer := range manifest.Layers {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(layer.Digest, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				c.JSON(500, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Unable to check the DB for package version",
						},
					},
				})
				return
			}
			if asset.Digest != layerDigest {
				c.JSON(404, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Uploaded asset not found",
						},
					},
				})
				return
			}
			versionSize += asset.Size
			assets = append(assets, asset)
		}
	case ManifestListV2ContentType, ManifestOCIIndexV1ContentType:
		manifest := ManifestListV2{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			c.JSON(400, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable tօ parse the manifest",
					},
				},
			})
			return
		}
		for _, manifestDescriptor := range manifest.Manifests {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(manifestDescriptor.Digest, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				c.JSON(500, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Unable to check the DB for package version",
						},
					},
				})
				return
			}
			if asset.Digest != layerDigest {
				c.JSON(404, gin.H{
					"errors": []gin.H{
						{
							"code":    "DENIED",
							"message": "authentication required",
							"detail":  "Uploaded asset not found",
						},
					},
				})
				return
			}

			versionSize += int64(manifestDescriptor.Size)
			assets = append(assets, asset)
		}
	default:
		c.JSON(400, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Bad Request",
				},
			},
		})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable tօ parse the manifest",
				},
			},
		})
		return
	}

	pkg := models.Package[PackageMetadata]{
		Namespace: authCtx.Namespace,
		Service:   s.Prefix,
	}
	err = pkg.FillByName(pkgName)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to check the DB for package",
				},
			},
		})
		return
	}

	if pkg.ID == uuid.Nil {
		pkg = models.Package[PackageMetadata]{
			Name:      pkgName,
			Service:   s.Prefix,
			AuthId:    authCtx.AuthId,
			Namespace: authCtx.Namespace,
		}
		err = pkg.Insert()
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable to create package",
					},
				},
			})
			return
		}
	}

	pkgVersion, err := pkg.Version(tagName)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to check the DB for package version",
				},
			},
		})
		return
	}
	if pkgVersion.ID == uuid.Nil {
		pkgVersion = models.PackageVersion[PackageMetadata]{
			PackageId: pkg.ID,
			AuthId:    authCtx.AuthId,
			Namespace: authCtx.Namespace,
			Service:   s.Prefix,
			Digest:    digest,
			Version:   tagName,
			Tag:       tagName,
			Size:      versionSize,
			Metadata:  datatypes.NewJSONType[PackageMetadata](metadata),
		}
		err = pkgVersion.Save()
		if err == nil {
			err = pkgVersion.SetAssets(assets)
		}
	} else {
		if pkgVersion.PackageId != pkg.ID || pkgVersion.Service != s.Prefix {
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
		pkgVersion.Version = tagName
		pkgVersion.Tag = tagName
		pkgVersion.Metadata = datatypes.NewJSONType[PackageMetadata](metadata)
		pkgVersion.Digest = metadata.Digest
		pkgVersion.Size = versionSize
		err = pkgVersion.Save()

		if err != nil {
			c.JSON(500, gin.H{
				"errors": []gin.H{
					{
						"code":    "DENIED",
						"message": "authentication required",
						"detail":  "Unable to insert package version",
					},
				},
			})
			return
		}

		err = pkgVersion.SetAssets(assets)
	}

	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "authentication required",
					"detail":  "Unable to insert package version",
				},
			},
		})
		return
	}

	pkg.LatestVersion = pkgVersion.Version
	err = pkg.Save()
	if err != nil {
		log.Println(err)
	}

	c.Header("Docker-Content-Digest", "sha256:"+digest)
	c.Status(201)
	c.Done()
}

type sizeHandler struct {
	size int64
}

type partialReadWriter struct {
	io.Reader
	input     io.Reader
	output    io.Writer
	totalSize *sizeHandler
}

func (p partialReadWriter) Read(b []byte) (n int, err error) {
	n, err = p.input.Read(b)
	p.totalSize.size += int64(n)
	_, _ = p.output.Write(b[:n])
	return n, err
}

func (s *Service) appendStorageData(uploadUUID string, input io.Reader) (digest string, size int64, err error) {
	fileReader, err := s.Storage.GetFile(s.PackageFilename(uploadUUID))
	if err != nil {
		return "", 0, err
	}

	var inputReader io.Reader

	if fileReader == nil {
		inputReader = input
	} else {
		inputReader = io.MultiReader(fileReader, input)
	}

	defer func(fileReader io.ReadCloser) {
		if fileReader == nil {
			return
		}
		_ = fileReader.Close()
	}(fileReader)

	hasher := sha256.New()

	sh := &sizeHandler{}
	rw := partialReadWriter{
		input:     inputReader,
		output:    hasher,
		totalSize: sh,
	}

	err = s.Storage.WriteFile(s.PackageFilename(uploadUUID), nil, rw)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(hasher.Sum(nil)), sh.size, nil
}
