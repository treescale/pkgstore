package middlewares

import (
	"github.com/gin-gonic/gin"
)

type AuthResult struct {
	PublicAccess bool   `json:"public_access"`
	Read         bool   `json:"read"`
	Write        bool   `json:"write"`
	Delete       bool   `json:"delete"`
	AuthId       string `json:"auth_id"`
	Namespace    string `json:"namespace"`
	Error        string `json:"error"`
}

func AuthMiddleware(c *gin.Context) {
	// Keeping namespace available for everyone by default
	c.Set("auth", &AuthResult{
		Read:         true,
		Write:        true,
		Delete:       true,
		PublicAccess: true,
		Namespace:    "",
		AuthId:       "public",
	})
	c.Next()
}

func GetAuthCtx(c *gin.Context) *AuthResult {
	return c.MustGet("auth").(*AuthResult)
}
