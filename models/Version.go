package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/treescale/pkgstore/db"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var digestRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)

type PackageVersion[MetaType any] struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey;" json:"id" binding:"required"`
	Service   string    `gorm:"column:service;not null" json:"service" binding:"required"`
	AuthId    string    `gorm:"column:auth_id;index;not null" json:"auth_id" binding:"required"`
	Namespace string    `gorm:"column:namespace;index;not null" json:"namespace" binding:"required"`

	Digest string `gorm:"column:digest;index" json:"digest"`
	Size   int64  `gorm:"column:size;not null" json:"size" binding:"required"`

	PackageId uuid.UUID `gorm:"column:package_id;uniqueIndex:pkg_id_version;uniqueIndex:pkg_id_tag;" json:"package_id" binding:"required"`

	Version string `gorm:"column:version;not null;uniqueIndex:pkg_id_version" json:"version" binding:"required"`
	Tag     string `gorm:"column:tag;uniqueIndex:pkg_id_tag" json:"tag"`

	Metadata datatypes.JSONType[MetaType] `gorm:"column:metadata" json:"metadata"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	AssetIds string `gorm:"column:asset_ids" json:"asset_ids"`
}

func (p *PackageVersion[T]) BeforeCreate(_ *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

func (*PackageVersion[T]) TableName() string {
	return "package_versions"
}

func (p *PackageVersion[T]) FillByName(version string) error {
	return db.DB().Order("created_at desc").Find(p, "version = ? AND namespace = ? AND service = ?", version, p.Namespace, p.Service).Preload("Asset").Error
}

func (p *PackageVersion[T]) FillById(id uint64) error {
	return db.DB().Find(p, "id = ? AND namespace = ", id, p.Namespace).Preload("Asset").Error
}

func (p *PackageVersion[T]) FillByDigest(digest string) error {
	match := digestRegex.MatchString(digest)
	if !match {
		return errors.New("invalid digest")
	}
	return db.DB().Order("created_at desc").Find(p, "digest = ? AND namespace = ? AND service = ?", digest, p.Namespace, p.Service).Preload("Asset").Error
}

func (p *PackageVersion[T]) Insert() error {
	return db.DB().Create(p).Error
}

func (p *PackageVersion[T]) SaveMeta() error {
	return db.DB().Model(p).Update("metadata", p.Metadata).Error
}

func (p *PackageVersion[T]) Save() error {
	return db.DB().Save(p).Error
}

func (p *PackageVersion[T]) Delete() error {
	return db.DB().Delete(&PackageVersion[T]{}, "id = ?", p.ID).Error
}

func (p *PackageVersion[T]) AddAsset(asset *Asset) error {
	if p.ID == uuid.Nil {
		return nil
	}
	id := asset.ID.String()
	if p.AssetIds == "" {
		p.AssetIds = id
	} else if !strings.Contains(p.AssetIds, id) {
		p.AssetIds = p.AssetIds + "," + id
	} else {
		return nil
	}
	return p.Save()
}

func (p *PackageVersion[T]) GetAssets() (assets []Asset, err error) {
	err = db.DB().Find(&assets, "id IN ?", strings.Split(p.AssetIds, ",")).Error
	return
}

func (p *PackageVersion[T]) SetAssets(assets []Asset) (err error) {
	for _, asset := range assets {
		if p.AssetIds == "" {
			p.AssetIds = asset.ID.String()
		} else if !strings.Contains(p.AssetIds, asset.ID.String()) {
			p.AssetIds = p.AssetIds + "," + asset.ID.String()
		}
	}
	return p.Save()
}
