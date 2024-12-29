package container

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/config"
	"github.com/treescale/pkgstore/models"
	"github.com/treescale/pkgstore/services"
	"github.com/treescale/pkgstore/storage"
)

const (
	ManifestV1ContentType         = "application/vnd.docker.distribution.manifest.v1+json"
	ManifestV2ContentType         = "application/vnd.docker.distribution.manifest.v2+json"
	ManifestListV2ContentType     = "application/vnd.docker.distribution.manifest.list.v2+json"
	ManifestOCIV1ContentType      = "application/vnd.oci.image.manifest.v1+json"
	ManifestOCIIndexV1ContentType = "application/vnd.oci.image.index.v1+json"
)

type PackageMetadata struct {
	ContentType    string `json:"contentType"`
	Digest         string `json:"digest"`
	MetadataBuffer []byte `json:"metadataBuffer"`
}

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{
			Prefix:                   "container",
			Storage:                  storage,
			PublicRegistryPathPrefix: "/v2/",
			PublicRegistryUrl:        "https://registry.hub.docker.com",
		},
	}
}

func (s *Service) PkgInfoFromRequest(c *gin.Context) (pkgName string, filename string) {
	pkgPath := c.Param("path")

	// /:pkgName/
	pattern := `^/v2/(?P<pkgName>([^/]+/)?[^/]+)/(blob/|manifest/)(?P<filename>[a-z0-9]+)(?:/)?$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(pkgPath)
	if matches == nil {
		return "", ""
	}

	for i, name := range re.SubexpNames() {
		if name == "pkgName" {
			pkgName = matches[i]
		} else if name == "filename" {
			filename = matches[i]
		}
	}

	return pkgName, filename
}

func (s *Service) SetAuthHeaderAndAbort(c *gin.Context, message string) {
	registryHost := config.Get().RegistryHosts.Container
	registryHostUrl, err := url.Parse(registryHost)
	if err != nil {
		c.JSON(500, gin.H{
			"errors": []gin.H{
				{
					"code":    "DENIED",
					"message": "Unable to parse registry host",
					"detail":  "authentication required",
				},
			},
		})
		c.Abort()
		return
	}
	c.Header("www-authenticate", fmt.Sprintf(`Bearer realm="%[1]s/",service="%[2]s"`, registryHost, registryHostUrl.Hostname()))
	c.JSON(401, gin.H{
		"errors": []gin.H{
			{
				"code":    "UNAUTHORIZED",
				"message": message,
				"detail":  message,
			},
		},
	})
	c.Abort()
}

func (s *Service) GetAssetsByManifest(metadata *PackageMetadata) (assets []models.Asset, err error) {
	assets = make([]models.Asset, 0)

	switch metadata.ContentType {
	case ManifestV1ContentType:
		manifest := ManifestV1{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			return
		}
		if len(manifest.FsLayers) == 0 {
			return
		}
		for _, layer := range manifest.FsLayers {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(layer.BlobSum, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				return
			}
			if asset.Digest != layerDigest {
				return
			}
			assets = append(assets, asset)
		}
	case ManifestV2ContentType, ManifestOCIV1ContentType:
		manifest := ManifestV2{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			return
		}
		for _, layer := range manifest.Layers {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(layer.Digest, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				return
			}
			if asset.Digest != layerDigest {
				return
			}
			assets = append(assets, asset)
		}
	case ManifestListV2ContentType, ManifestOCIIndexV1ContentType:
		manifest := ManifestListV2{}
		err = json.Unmarshal(metadata.MetadataBuffer, &manifest)
		if err != nil {
			return
		}
		for _, manifestDescriptor := range manifest.Manifests {
			asset := models.Asset{
				Service: s.Prefix,
			}
			layerDigest := strings.Replace(manifestDescriptor.Digest, "sha256:", "", 1)
			err = asset.FillByDigest(layerDigest)
			if err != nil {
				return
			}
			if asset.Digest != layerDigest {
				return
			}

			assets = append(assets, asset)
		}
	}

	return
}

type ManifestV1 struct {
	Name     string `json:"name,omitempty"`
	Tag      string `json:"tag,omitempty"`
	FsLayers []struct {
		BlobSum string `json:"blobSum,omitempty"`
	} `json:"fsLayers,omitempty"`
	History []struct {
		V1Compatibility string `json:"v1Compatibility,omitempty"`
	} `json:"history,omitempty"`
	SchemaVersion int `json:"schemaVersion,omitempty"`
	Signatures    []struct {
		Header struct {
			Jwk struct {
				Crv string `json:"crv,omitempty"`
				Kid string `json:"kid,omitempty"`
				Kty string `json:"kty,omitempty"`
				X   string `json:"x,omitempty"`
				Y   string `json:"y,omitempty"`
			} `json:"jwk"`
			Alg string `json:"alg,omitempty"`
		} `json:"header,omitempty"`
		Signature string `json:"signature,omitempty"`
		Protected string `json:"protected,omitempty"`
	} `json:"signatures,omitempty"`
}

type ManifestListV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"` // application/vnd.docker.distribution.manifest.list.v2+json
	ArtifactType  string `json:"artifactType,omitempty"`
	Manifests     []struct {
		MediaType string `json:"mediaType"` // application/vnd.docker.distribution.manifest.v2+json
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
		Platform  struct {
			Architecture string   `json:"architecture,omitempty"`
			Os           string   `json:"os,omitempty"`
			OsVersion    string   `json:"os.version,omitempty"`
			OsFeatures   []string `json:"os.features,omitempty"`
			Variant      string   `json:"variant,omitempty"`
			Features     []string `json:"features,omitempty"`
		} `json:"platform,omitempty"`
	} `json:"manifests,omitempty"`
	Subject struct {
		MediaType    string            `json:"mediaType,omitempty"`
		Size         int               `json:"size,omitempty"`
		Digest       string            `json:"digest,omitempty"`
		Urls         []string          `json:"urls,omitempty"`
		Annotations  map[string]string `json:"annotations,omitempty"`
		Data         string            `json:"data,omitempty"`
		ArtifactType string            `json:"artifactType,omitempty"`
	} `json:"subject,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ManifestV2 struct {
	SchemaVersion int    `json:"schemaVersion,omitempty"`
	MediaType     string `json:"mediaType,omitempty"`
	Config        struct {
		MediaType string `json:"mediaType,omitempty"`
		Size      int    `json:"size,omitempty"`
		Digest    string `json:"digest,omitempty"`
	} `json:"config"`
	Layers []struct {
		MediaType   string            `json:"mediaType,omitempty"`
		Size        int               `json:"size,omitempty"`
		Digest      string            `json:"digest,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
	} `json:"layers"`
}
