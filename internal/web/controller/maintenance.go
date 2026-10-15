package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

// Maintenance routes inherit API authentication, token scope enforcement and CSRF.
func registerMaintenanceRoutes(api *gin.RouterGroup) {
	g := api.Group("/maintenance")
	settings := &service.SettingService{}
	server := &service.ServerService{}
	g.GET("/backups", func(c *gin.Context) {
		value, err := settings.GetFullBackupSettings()
		if err != nil {
			jsonObj(c, nil, err)
			return
		}
		status, err := settings.GetFullBackupStatus()
		if err != nil {
			jsonObj(c, nil, err)
			return
		}
		files, err := service.ListFullBackups()
		jsonObj(c, gin.H{"settings": value, "status": status, "files": files}, err)
	})
	g.POST("/backups/settings", func(c *gin.Context) {
		var value service.FullBackupSettings
		if err := c.ShouldBindJSON(&value); err != nil {
			jsonObj(c, nil, err)
			return
		}
		jsonObj(c, nil, settings.SaveFullBackupSettings(value))
	})
	g.POST("/backups/run", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
		defer cancel()
		status, err := server.RunFullBackup(ctx, false)
		jsonObj(c, status, err)
	})
	g.GET("/backups/download/:name", func(c *gin.Context) {
		path, err := service.FullBackupPath(c.Param("name"))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "no-store")
		c.FileAttachment(path, c.Param("name"))
	})
	g.GET("/nginx", func(c *gin.Context) {
		value, err := settings.GetNginxSitePaths()
		if err != nil {
			jsonObj(c, nil, err)
			return
		}
		if path := c.Query("configPath"); path != "" {
			value.ConfigPath = path
		}
		if root := c.Query("root"); root != "" {
			value.Root = root
		}
		value, err = settings.ReadNginxSite(value)
		jsonObj(c, value, err)
	})
	g.GET("/nginx/sites", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
		defer cancel()
		value, err := service.DiscoverNginxSites(ctx)
		jsonObj(c, value, err)
	})
	g.POST("/nginx", func(c *gin.Context) {
		var value service.NginxSite
		if err := c.ShouldBindJSON(&value); err != nil {
			jsonObj(c, nil, err)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		jsonObj(c, nil, settings.ApplyNginxSite(ctx, value))
	})
}

func scheduledFullBackup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if _, err := (&service.ServerService{}).RunFullBackup(ctx, true); err != nil {
		logger.Warning("scheduled full backup: ", err)
	}
}
