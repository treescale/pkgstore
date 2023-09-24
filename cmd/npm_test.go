package cmd

import (
	"github.com/alin-io/pkgproxy/db"
	"github.com/alin-io/pkgproxy/models"
	"github.com/alin-io/pkgproxy/router"
	"github.com/alin-io/pkgproxy/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNPMUploadHandler(t *testing.T) {
	storageBackend := storage.NewInMemoryBackend()
	db.InitDatabaseForTest()
	models.SyncModels()

	r := router.SetupGinServer(storageBackend)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}
