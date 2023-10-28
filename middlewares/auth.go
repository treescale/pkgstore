package middlewares

import (
	"encoding/base64"
	"fmt"
	"github.com/alin-io/pkgstore/config"
	"github.com/carlmjohnson/requests"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"log"
	"strings"
	"time"
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

const AUTH_ID_PUBLIC = "public"

var (
	// make cache with 10s TTL and 1000 max keys
	authCache = expirable.NewLRU[string, *AuthResult](1000, nil, time.Second*10)
)

func AuthMiddleware(c *gin.Context) {
	if len(config.Get().AuthEndpoint) == 0 {
		c.Set("auth", &AuthResult{
			PublicAccess: true,
			Read:         true,
			Write:        true,
			Delete:       true,
			Namespace:    "",
			AuthId:       "public",
		})
		c.Next()
		return
	}

	requestService := getServiceFromPath(c)
	tokenHeader := c.GetHeader("Authorization")

	pkgAction := "pull"
	if c.Request.Method == "PUT" || c.Request.Method == "POST" || c.Request.Method == "PATCH" {
		pkgAction = "push"
	}

	if len(tokenHeader) == 0 {
		_, tokenHeader, _ = c.Request.BasicAuth()
	}

	authResult := &AuthResult{
		PublicAccess: true,
		AuthId:       AUTH_ID_PUBLIC,
	}

	if len(tokenHeader) > 0 {
		tokenSplit := strings.Split(tokenHeader, " ")
		if len(tokenSplit) != 2 {
			c.AbortWithStatus(401)
			return
		}
		tokenString := tokenSplit[1]
		decodedToken, err := decodeBase64ToUnicode(tokenString)
		if err == nil {
			tokenString = decodedToken
		}

		cacheKey := fmt.Sprintf("%s-%s-%s", tokenString, requestService, pkgAction)

		if r, ok := authCache.Get(cacheKey); ok {
			authResult = r
		} else {
			err := requests.URL(config.Get().AuthEndpoint).
				Header("Authorization", tokenString).
				Header("X-Package-Service", requestService).
				Header("X-Package-Action", pkgAction).
				ToJSON(authResult).
				ErrorJSON(&authResult).
				Fetch(c)
			if err != nil || authResult.Error != "" {
				log.Println(err)
				c.JSON(401, gin.H{
					"errors": []gin.H{
						{
							"code":    "UNAUTHORIZED",
							"message": authResult.Error,
						},
					},
				})
				c.Abort()
				return
			}

			authCache.Add(cacheKey, authResult)
		}
	}

	c.Set("auth", authResult)
	c.Next()
}

func getServiceFromPath(c *gin.Context) string {
	fullPath := c.FullPath()

	hostname := c.Request.Host
	hostnameSplit := strings.Split(hostname, ".")
	if len(hostnameSplit) == 3 {
		return hostnameSplit[0]
	}

	pathSplit := strings.Split(fullPath, "/")
	if len(pathSplit) < 2 {
		return ""
	}
	pathPrefix := pathSplit[1]
	if pathPrefix == "v2" {
		return "container"
	}
	return pathPrefix
}

func GetAuthCtx(c *gin.Context) *AuthResult {
	return c.MustGet("auth").(*AuthResult)
}

func decodeBase64ToUnicode(str string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	utf8String := string(decodedBytes)

	decodedUnicode := strings.ToValidUTF8(utf8String, "")
	if utf8String != decodedUnicode {
		return "", fmt.Errorf("the decoded string contains invalid UTF-8 characters")
	}

	return decodedUnicode, nil
}
