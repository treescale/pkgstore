package db

import (
	"github.com/alin-io/pkgstore/config"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
)

var (
	client *gorm.DB
)

func InitDatabase() {
	var err error
	var dialector gorm.Dialector

	if strings.Index(config.Get().DatabaseUrl, "postgres://") == 0 {
		dialector = postgres.Open(config.Get().DatabaseUrl)
	} else {
		dialector = sqlite.Open(config.Get().DatabaseUrl)
	}

	loggingMode := logger.Info
	if gin.Mode() == gin.ReleaseMode {
		loggingMode = logger.Error
	}

	client, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(loggingMode),
	})
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
