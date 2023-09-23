package models

import (
	"gorm.io/gorm"
)

type Package[MetaType any] struct {
	gorm.Model

	Id            uint64                     `gorm:"column:id;primaryKey;autoincrement" json:"id" binding:"required"`
	Name          string                     `gorm:"column:name;unique;not null" json:"name" binding:"required"`
	Namespace     string                     `gorm:"column:namespace;index;not null" json:"namespace" binding:"required"`
	Service       string                     `gorm:"column:service;index;not null" json:"service" binding:"required"`
	AuthId        string                     `gorm:"column:auth_id;index;not null" json:"auth_id" binding:"required"`
	LatestVersion string                     `gorm:"column:latest_version" json:"latest_version"`
	Versions      []PackageVersion[MetaType] `gorm:"foreignKey:PackageId;references:Id" json:"versions"`
}
