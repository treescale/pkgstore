package models

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/alin-io/pkgstore/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Asset struct {
	ID      uuid.UUID `gorm:"column:id;primaryKey;" json:"id" binding:"required"`
	Service string    `gorm:"column:service;not null" json:"service" binding:"required"`

	Digest string `gorm:"column:digest;index,not null;uniqueIndex" json:"digest" binding:"required"`
	Size   int64  `gorm:"column:size;not null" json:"size" binding:"required"`

	UploadUUID  string `gorm:"column:upload_uuid;uniqueIndex;not null" json:"upload_uuid" binding:"required"`
	UploadRange string `gorm:"column:upload_range;not null" json:"upload_range" binding:"required"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (t *Asset) BeforeCreate(_ *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (*Asset) TableName() string {
	return "assets"
}

func (t *Asset) StartUpload() error {
	t.UploadUUID = uuid.NewString()
	t.UploadRange = "0-0"
	t.SetRandomDigest()
	return db.DB().Create(t).Error
}

func (t *Asset) Insert() error {
	err := db.DB().Create(t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (t *Asset) FillByDigest(digest string) error {
	match := digestRegex.MatchString(digest)
	if !match {
		return errors.New("invalid digest")
	}
	return db.DB().Find(t, `digest = ? AND service = ?`, digest, t.Service).Error
}

func (t *Asset) FillById(id string) error {
	return db.DB().Find(t, "id = ? AND service = ?", id, t.Service).Error
}

func (t *Asset) FillByUploadUUID(uploadUUID string) error {
	return db.DB().Find(t, `"upload_uuid" = ? AND service = ?`, uploadUUID, t.Service).Error
}

func (t *Asset) Update() error {
	return db.DB().Save(t).Error
}

func (t *Asset) Delete() error {
	return db.DB().Delete(t).Error
}

func (t *Asset) SetRandomDigest() {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	hash := sha256.Sum256(data)
	t.Digest = fmt.Sprintf("%x", hash)
}

func (t *Asset) GetVersion() (version *PackageVersion[any], err error) {
	version = &PackageVersion[any]{}
	err = db.DB().Find(version, `asset_ids LIKE ?`, "%"+t.ID.String()+"%").Error
	return

}
