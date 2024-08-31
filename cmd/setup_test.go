package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/treescale/pkgstore/db"
	"github.com/treescale/pkgstore/models"
	"github.com/treescale/pkgstore/router"
	"github.com/treescale/pkgstore/services/npm"
	"github.com/treescale/pkgstore/storage"
)

var (
	storageBackend storage.BaseStorageBackend
	serverApp      *gin.Engine
)

func init() {
	storageBackend = storage.NewInMemoryBackend()
	db.InitDatabaseForTest()
	models.SyncModels()

	serverApp = router.SetupGinServer()
	serverApp.Use(func(c *gin.Context) {
		c.Set("testing", true)
	})
	router.PackageRouter(serverApp, storageBackend)
}

func TestServerHealthCheck(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	serverApp.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func DeleteTestPackage(name, service string) error {
	return db.DB().Delete(&models.Package[npm.MetadataResponse]{}, "name = ? AND service = ?", name, service).Error
}
