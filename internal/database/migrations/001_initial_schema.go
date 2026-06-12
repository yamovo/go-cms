package migrations

import (
	"github.com/vortexcms/go-cms/internal/database"
	"github.com/vortexcms/go-cms/internal/models"
	"gorm.io/gorm"
)

func init() {
	RegisterMigrations(
		database.Migration{
			Version:     1,
			Description: "Create initial schema",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					// Auth & Users
					&models.Permission{},
					&models.Role{},
					&models.User{},

					// Content
					&models.Category{},
					&models.Tag{},
					&models.Article{},
					&models.Comment{},
					&models.Revision{},
					&models.CustomField{},

					// Media
					&models.Media{},

					// Navigation
					&models.Menu{},
					&models.MenuItem{},

					// Settings & SEO
					&models.SiteSetting{},
					&models.SEOSetting{},
					&models.RedirectRule{},

					// Extensions
					&models.Plugin{},

					// Analytics
					&models.PageView{},
					&models.ActivityLog{},
				)
			},
			Down: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&models.ActivityLog{},
					&models.PageView{},
					&models.Plugin{},
					&models.RedirectRule{},
					&models.SEOSetting{},
					&models.SiteSetting{},
					&models.MenuItem{},
					&models.Menu{},
					&models.Media{},
					&models.CustomField{},
					&models.Revision{},
					&models.Comment{},
					&models.Article{},
					&models.Tag{},
					&models.Category{},
					&models.User{},
					&models.Role{},
					&models.Permission{},
				)
			},
		},
	)
}
