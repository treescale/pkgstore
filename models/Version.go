package models

import (
	"github.com/alin-io/pkgproxy/db"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PackageVersion[MetaType any] struct {
	gorm.Model

	Digest string `gorm:"column:digest;primaryKey" json:"digest" binding:"required"`

	PackageId uint64 `gorm:"column:package_id" json:"package_id" binding:"required"`

	Version string `gorm:"column:version;not null" json:"version" binding:"required"`
	Tag     string `gorm:"column:tag" json:"tag"`

	Size uint64 `gorm:"column:size;not null" json:"size" binding:"required"`

	Metadata datatypes.JSONType[MetaType] `gorm:"column:metadata" json:"metadata"`
}

func (*PackageVersion[T]) TableName() string {
	return "package_versions"
}

func (p *PackageVersion[T]) SaveMeta() error {
	return db.DB().Model(p).Update("metadata", p.Metadata).Error
}

func (p *PackageVersion[T]) Delete() error {
	return db.DB().Delete(&PackageVersion[T]{}, "digest = ?", p.Digest).Error
}
