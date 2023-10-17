package config

import (
	// Autoload .env file
	_ "github.com/joho/godotenv/autoload"

	"os"
)

var (
	projectConfig = ProjectConfigType{}
)

const (
	StorageS3         = "s3"
	StorageFileSystem = "filesystem"

	// NumberOfPkgNameLevels PkgName Levels (e.g. /npm/@username/package-name)
	NumberOfPkgNameLevels = 2
)

func init() {
	projectConfig.Inti()
}

func Get() *ProjectConfigType {
	return &projectConfig
}

type ProjectConfigType struct {
	ListenAddress string
	DatabaseUrl   string
	RegistryHost  string
	Storage       struct {
		ActiveBackend  string
		FileSystemRoot string
		S3             struct {
			Region    string
			Bucket    string
			ApiKey    string
			ApiSecret string
			ApiHost   string
		}
	}
}

func (c *ProjectConfigType) Inti() {
	c.ListenAddress = GetEnv("LISTEN_ADDRESS", ":8080")
	c.RegistryHost = GetEnv("REGISTRY_HOST", "http://localhost:8080")
	c.DatabaseUrl = GetEnv("DATABASE_URL", "file::memory:?cache=shared")

	// Storage Backend
	c.Storage.ActiveBackend = GetEnv("STORAGE_BACKEND", StorageS3)

	// S3 Storage Config
	c.Storage.S3.Region = GetEnv("S3_REGION", "us-east-1")
	c.Storage.S3.Bucket = GetEnv("S3_BUCKET", "pkgstore")
	c.Storage.S3.ApiKey = GetEnv("S3_API_KEY", "minioadmin")
	c.Storage.S3.ApiSecret = GetEnv("S3_API_SECRET", "minioadmin")
	c.Storage.S3.ApiHost = GetEnv("S3_API_HOST", "")

	// File System Storage Config
	c.Storage.FileSystemRoot = GetEnv("STORAGE_BACKEND_FILESYSTEM_ROOT", "")
}

func GetEnv(keyAndFallback ...string) string {
	key := keyAndFallback[0]
	value := os.Getenv(key)
	if len(value) == 0 {
		if len(keyAndFallback) > 1 {
			return keyAndFallback[1]
		}
		panic("Missing environment variable - " + key)
	}
	return value
}
