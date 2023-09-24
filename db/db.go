package db

import (
	"github.com/alin-io/pkgproxy/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	client *gorm.DB
)

func InitDatabase() {
	var err error
	client, err = gorm.Open(sqlite.Open(config.Get().DatabaseUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func InitDatabaseForTest() {
	var err error
	client, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func DB() *gorm.DB {
	return client
}
