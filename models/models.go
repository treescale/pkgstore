package models

import (
	"github.com/treescale/pkgstore/db"
)

func SyncModels() {
	err := db.DB().AutoMigrate(&Package[any]{}, &PackageVersion[any]{}, &Asset{})
	if err != nil {
		panic(err)
	}
}
