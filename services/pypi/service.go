package pypi

import (
	"github.com/treescale/pkgstore/services"
	"github.com/treescale/pkgstore/storage"
)

type PackageMetadata struct {
	RequiresPython string   `json:"requires_python"`
	OriginalFiles  []string `json:"original_files"`
}

type Service struct {
	services.BasePackageService
}

func NewService(storage storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{
			Prefix:                   "pypi",
			Storage:                  storage,
			PublicRegistryPathPrefix: "/simple/",
			PublicRegistryUrl:        "https://pypi.org",
		},
	}
}
