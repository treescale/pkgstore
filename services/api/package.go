package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/treescale/pkgstore/db"
	"github.com/treescale/pkgstore/middlewares"
	"github.com/treescale/pkgstore/models"
)

func (s *Service) ListPackagesHandler(c *gin.Context) {
	nameFilter := c.Query("q")
	pkgs := make([]models.Package[any], 0)
	err := db.DB().Model(&pkgs).Where("name LIKE ? AND auth_id = ?", "%"+nameFilter+"%", middlewares.GetAuthCtx(c).AuthId).Preload("Versions").Find(&pkgs).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, pkgs)
}

func (s *Service) GetPackage(c *gin.Context) {
	packageIdString := c.Param("id")
	packageId, err := uuid.Parse(packageIdString)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid package id"})
		return
	}

	pkg := models.Package[any]{}
	err = db.DB().Model(&pkg).Preload("Versions").Where(`id = ? AND auth_id = ?`, packageId, middlewares.GetAuthCtx(c).AuthId).Find(&pkg).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if pkg.ID == uuid.Nil {
		c.JSON(404, gin.H{"error": "Package not found"})
		return
	}
	c.JSON(200, pkg)
}

func (s *Service) DeletePackage(c *gin.Context) {
	packageIdString := c.Param("id")
	packageId, err := uuid.Parse(packageIdString)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid package id"})
		return
	}

	pkg := models.Package[any]{}
	err = db.DB().Model(&pkg).Where("id = ? AND auth_id = ?", packageId, middlewares.GetAuthCtx(c).AuthId).Find(&pkg).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if pkg.ID == uuid.Nil {
		c.JSON(404, gin.H{"error": "Package not found"})
		return
	}
	err = pkg.Delete()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, pkg)
}
