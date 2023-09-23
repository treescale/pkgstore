package pypi

import (
	"fmt"
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/storage"
)

type pypiPackageMetadata struct {
	RequiresPython  string `json:"requires_python"`
	FilenamePostfix string `json:"filename_postfix"`
}

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{Prefix: "pypi", Storage: storage},
	}
}

func (s *Service) constructPackageOriginalFilename(name, version, postfix string) string {
	if len(postfix) > 0 {
		postfix = "-" + postfix
	}
	return fmt.Sprintf("%s-%s%s", name, version, postfix)
}

func (s *Service) PypiPackageFilename(digest, postfix string) string {
	return fmt.Sprintf("%s/%s%s", s.Prefix, digest, postfix)
}
