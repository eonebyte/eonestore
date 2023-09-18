package handler

import (
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"gorm.io/gorm"
	"github.com/gofiber/fiber/v2"
)

func CheckDatabaseInstallation(c *fiber.Ctx) error {

	var setting models.Setting
	err := config.DB.Last(&setting, "installed = ?", "true").Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if c.Path() == "/" {
				return c.Redirect("/install")
			}
			if c.Path() == "/install" || c.Path() == "/handle-install" {
				return c.Next()
			}
			return c.Redirect("/install")
		}
	}
	if setting.Installed == "true" {
		if c.Path() == "/install" || c.Path() == "/handle-install" {
			return c.Redirect("/auth")
		}
		return c.Next()
	}
	
	return c.Next()
}