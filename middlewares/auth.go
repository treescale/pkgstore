package middlewares

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"strings"
)

func AuthMiddleware(c *gin.Context) {
	tokenString := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1)
	if strings.Contains(tokenString, "Basic ") {
		tokenStringBytes, err := base64.StdEncoding.DecodeString(strings.Replace(tokenString, "Basic ", "", 1))
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		tokenString = string(tokenStringBytes)
	}
	tokenSplit := strings.Split(tokenString, ":")
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
