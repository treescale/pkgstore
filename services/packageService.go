package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/alin-io/pkgproxy/storage"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

type PackageService interface {
	PackageFilename(digest, postfix string) string
	PkgVersionFromFilename(filename string) (pkgName string, version string)
	ShouldHandleRequest(c *gin.Context) bool
	PkgInfoFromRequestPath(c *gin.Context) (pkgName string, filename string)

	UploadHandler(c *gin.Context)
	DownloadHandler(c *gin.Context)
	MetadataHandler(c *gin.Context)
}

type BasePackageService struct {
	PackageService

	Prefix  string
	Storage storage.BaseStorageBackend
}

func (s *BasePackageService) PackageFilename(digest, postfix string) string {
	if len(postfix) == 0 {
		postfix = ".tar.gz"
	}
	return fmt.Sprintf("%s/%s%s", s.Prefix, digest, postfix)
}

func (s *BasePackageService) PkgVersionFromFilename(filename string) (pkgName string, version string) {
	filenameSplit := strings.Split(filename, "-")
	pkgName = filenameSplit[0]
	version = strings.Replace(filenameSplit[1], ".tgz", "", 1)
	return pkgName, version
}

func (s *BasePackageService) ChecksumReader(r io.Reader) (checksum string, size int64, err error) {
	h := sha256.New()
	if size, err = io.Copy(h, r); err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), size, nil
}

func (s *BasePackageService) ShouldHandleRequest(c *gin.Context) bool {
	return c.GetString("pkgType") == s.Prefix
}
