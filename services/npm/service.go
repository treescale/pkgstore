package npm

import (
	"fmt"
	"github.com/alin-io/pkgproxy/storage"
	"strings"
)

type Service struct {
	storage storage.BaseStorageBackend
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) PackageFilename(digest string) string {
	return fmt.Sprintf("npm/%s.tgz", digest)
}

func (s *Service) PkgVersionFromFilename(filename string) (pkgName string, version string) {
	filenameSplit := strings.Split(filename, "-")
	pkgName = filenameSplit[0]
	version = strings.Replace(filenameSplit[1], ".tgz", "", 1)
	return pkgName, version
}
