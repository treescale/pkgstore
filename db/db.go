package db

import (
	"github.com/alin-io/pkgproxy/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	client *gorm.DB
)

func init() {
	var err error
	client, err = gorm.Open(sqlite.Open(config.Get().DatabaseUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func DB() *gorm.DB {
	return client
}
