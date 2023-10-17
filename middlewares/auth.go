package middlewares

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	c.Set("AuthId", "auth-token")

	// Full access without auth by default
	c.Set("AuthAccess", "read,write")

	// Keeping namespace available for everyone by default
	c.Set("AuthNamespace", "")
	c.Next()
}

func GetAuthId(c *gin.Context) string {
	return c.GetString("AuthId")
}
