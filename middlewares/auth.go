package middlewares

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	c.Set("AuthId", "auth-token")
	c.Next()
}

func GetAuthId(c *gin.Context) string {
	return c.GetString("AuthId")
}
