package middlewares

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func AuthMiddleware(c *gin.Context) {
	tokenSplit := strings.Split(strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1), ":")
	username := ""
	token := ""
	isBasicAuth := false
	if len(tokenSplit) == 2 {
		username = tokenSplit[0]
		token = tokenSplit[1]
	} else if len(tokenSplit) == 1 {
		token = tokenSplit[0]
	} else {
		username, token, isBasicAuth = c.Request.BasicAuth()
		if !isBasicAuth {
			c.AbortWithStatus(401)
			return
		}
	}

	c.Set("pkgType", username)
	c.Set("token", token)
	c.Next()
}
