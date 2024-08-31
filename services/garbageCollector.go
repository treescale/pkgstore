package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/treescale/pkgstore/db"
	"github.com/treescale/pkgstore/models"
	"github.com/treescale/pkgstore/storage"
)

type GarbageCollector struct {
	Storage storage.BaseStorageBackend
}

func (g *GarbageCollector) CleanupAssets(dryrun bool) (assets []models.Asset, err error) {
	tmpAssets := make([]models.Asset, 0)
	err = db.DB().Find(&tmpAssets).Error
	if err != nil {
		return
	}

	for _, asset := range tmpAssets {
		version, err := asset.GetVersion()
		if err != nil {
			return nil, err
		}

		if version == nil || version.ID == uuid.Nil {
			assets = append(assets, asset)
			if !dryrun {
				err = g.DeleteAsset(&asset)
				if err != nil {
					log.Println("Error while deleting asset", asset.ID, err)
				}
			}
		}
	}

	return
}

func (g *GarbageCollector) DeleteAsset(asset *models.Asset) (err error) {
	service := BasePackageService{
		Storage: g.Storage,
		Prefix:  asset.Service,
	}
	err = asset.Delete()
	if err != nil {
		return
	}

	err = g.Storage.DeleteFile(service.PackageFilename(asset.Digest))
	return err
}
