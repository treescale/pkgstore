package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/alin-io/pkgstore/config"
	"github.com/alin-io/pkgstore/storage"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
)

type PackageService interface {
	PackageFilename(digest string) string
	PkgVersionFromFilename(filename string) (pkgName string, version string)
	ConstructFullPkgName(c *gin.Context) (pkgName string)

	UploadHandler(c *gin.Context)
	DownloadHandler(c *gin.Context)
	MetadataHandler(c *gin.Context)

	SetAuthHeaderAndAbort(c *gin.Context)
	GetPrefix() string
}

type BasePackageService struct {
	PackageService

	Prefix  string
	Storage storage.BaseStorageBackend

	PublicRegistryUrl        string
	PublicRegistryPathPrefix string
}

func (s *BasePackageService) PackageFilename(digest string) string {
	return fmt.Sprintf("%s/%s", s.Prefix, digest)
}

func (s *BasePackageService) PkgVersionFromFilename(filename string) (pkgName string, version string) {
	base := filepath.Base(filename)
	for _, ext := range []string{".tar.gz", ".tgz", ".whl"} {
		if strings.HasSuffix(base, ext) {
			base = strings.Replace(base, ext, "", 1)
		}
	}

	filenameSplit := strings.Split(base, "-")
	pkgName = strings.Join(filenameSplit[:len(filenameSplit)-1], "-")
	version = filenameSplit[len(filenameSplit)-1]
	return pkgName, version
}

func (s *BasePackageService) ChecksumReader(r io.Reader) (checksum string, size int64, err error) {
	h := sha256.New()
	if size, err = io.Copy(h, r); err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), size, nil
}

func (s *BasePackageService) ConstructFullPkgName(c *gin.Context) string {
	pkgName := ""
	for i := 0; i < config.NumberOfPkgNameLevels; i++ {
		pkgParam := c.Param(fmt.Sprintf("name%d", i))
		if len(pkgParam) > 0 {
			pkgName = fmt.Sprintf("%s/%s", pkgName, pkgParam)
		}
	}
	if len(pkgName) == 0 {
		return ""
	}

	return pkgName[1:]
}

func (s *BasePackageService) ProxyToPublicRegistry(c *gin.Context) {
	urlPath := s.PublicRegistryUrl + c.Param("path")
	remote, err := url.Parse(s.PublicRegistryPathPrefix + urlPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header

		// Remove Authorization header
		req.Header.Del("Authorization")

		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = urlPath
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (s *BasePackageService) SetAuthHeaderAndAbort(c *gin.Context) {
	c.AbortWithStatus(401)
}

func (s *BasePackageService) GetPrefix() string {
	return s.Prefix
}
