package pypi

import "github.com/gin-gonic/gin"

func (s *Service) UploadHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
