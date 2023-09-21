package models

import "github.com/alin-io/pkgproxy/db"

func SyncModels() {
	err := db.DB().AutoMigrate(&Package{}, &PackageVersion{})
	if err != nil {
		panic(err)
	}
}
