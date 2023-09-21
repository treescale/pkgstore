package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PackageVersion struct {
	gorm.Model

	Digest string `gorm:"column:digest;primaryKey" json:"digest" binding:"required"`

	PackageId uint64 `gorm:"column:package_id" json:"package_id" binding:"required"`

	Version string `gorm:"column:version;not null" json:"version" binding:"required"`
	Tag     string `gorm:"column:tag" json:"tag"`

	Size uint64 `gorm:"column:size;not null" json:"size" binding:"required"`

	Metadata datatypes.JSONType[PackageVersionMetadata] `gorm:"column:metadata" json:"metadata"`
}

type PackageVersionMetadata struct {
	Id          string            `json:"_id"`
	Description string            `json:"description"`
	Readme      string            `json:"readme"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	NodeVersion string            `json:"_nodeVersion"`
	NpmVersion  string            `json:"_npmVersion"`
	Author      map[string]string `json:"author"`
	Dist        struct {
		Integrity string `json:"integrity"`
		Shasum    string `json:"shasum"`
		Tarball   string `json:"tarball"`
	} `json:"dist"`
	PublishConfig map[string]string `json:"publishConfig"`
	Scripts       map[string]string `json:"scripts"`
	Keywords      []string          `json:"keywords"`
	License       string            `json:"license"`
	Main          string            `json:"main"`
}
