package middlewares

import (
	"encoding/base64"
	"fmt"
	"github.com/alin-io/pkgstore/config"
	"github.com/carlmjohnson/requests"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
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

const AuthIdPublic = "public"

var (
	// make cache with 10s TTL and 1000 max keys
	authCache = expirable.NewLRU[string, *AuthResult](1000, nil, time.Second*10)
)

func extractTokenHeader(c *gin.Context) (string, error) {
	tokenHeader := c.GetHeader("Authorization")

	if len(tokenHeader) == 0 {
		_, tokenHeader, _ = c.Request.BasicAuth()
	}

	tokenSplit := strings.Split(tokenHeader, " ")
	tokenString := ""
	if len(tokenSplit) == 1 {
		tokenString = tokenSplit[0]
	} else if len(tokenSplit) == 2 {
		tokenString = tokenSplit[1]
	} else {
		return "", fmt.Errorf("invalid authorization header")
	}

	decodedToken, err := decodeBase64ToUnicode(tokenString)
	if err == nil {
		tokenString = decodedToken
	}

	return tokenString, nil
}

func getRemoteAuthContext(c *gin.Context, pkgName, token, pkgService, action string) (authResult *AuthResult, err error) {
	cacheKey := fmt.Sprintf("%s-%s-%s-%s", token, pkgName, pkgService, action)

	if r, ok := authCache.Get(cacheKey); ok {
		authResult = r
	} else {
		err := requests.URL(config.Get().AuthEndpoint).
			Header("Authorization", token).
			Header("X-Package-Service", pkgService).
			Header("X-Package-Name", pkgName).
			Header("X-Package-Action", action).
			ToJSON(authResult).
			ErrorJSON(&authResult).
			Fetch(c)
		if err != nil {
			return nil, err
		}

		if authResult.Error != "" {
			return authResult, nil
		}

		authCache.Add(cacheKey, authResult)
	}

	return authResult, nil
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
