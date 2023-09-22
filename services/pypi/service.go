package pypi

import (
	"fmt"
	"github.com/alin-io/pkgproxy/services"
	"github.com/alin-io/pkgproxy/storage"
)

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{Prefix: "pypi", Storage: storage},
	}
}

func (s *Service) constructPackageOriginalFilename(name, version, postfix string) string {
	return fmt.Sprintf("%s-%s-%s", name, version, postfix)
}
