package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/treescale/pkgstore/db"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/services"
	"github.com/treescale/pkgstore/storage"
)

type Service struct {
	services.BasePackageService
}

type RegistryStatsResponse struct {
	NumPackages int `json:"num_packages"`
	NumVersions int `json:"num_versions"`
	StorageSize int `json:"storage_size"`
}

func NewApiService(storageBackend storage.BaseStorageBackend) *Service {
	return &Service{
		BasePackageService: services.BasePackageService{
			Storage: storageBackend,
			Prefix:  "api",
		},
	}
}

func (s *Service) RegistryStats(c *gin.Context) {
	authId := middlewares.GetAuthCtx(c).AuthId
	result := RegistryStatsResponse{}
	err := db.DB().Raw(`SELECT COUNT(*) AS num_packages,
       (SELECT COUNT(*) FROM package_versions WHERE auth_id = @auth_id) AS num_versions,
       (SELECT SUM(size) FROM package_versions WHERE auth_id = @auth_id) AS storage_size
FROM packages WHERE auth_id = @auth_id`, sql.Named("auth_id", authId)).Scan(&result).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to get stats"})
		return
	}
	c.JSON(200, result)
}
