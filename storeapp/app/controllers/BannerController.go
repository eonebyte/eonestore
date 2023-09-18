package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"os"
	"path/filepath"
	"time"
)

type BannerController struct {
	baseUrl string
	session *session.Store
}

func NewBannerController(session *session.Store, b string) *BannerController {
	return &BannerController{
		session: session,
		baseUrl: b,
	}
}

func (b *BannerController) PageBanner(c *fiber.Ctx) error {
	session, err := b.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}

	//get Session Data before Decode
	sessionDataJSON := session.Get("sessionData")

	// Convert the JSON data back to the User struct
	var sessionData SessionData
	err = json.Unmarshal(sessionDataJSON.([]byte), &sessionData)
	if err != nil {
		return c.SendString("Failed to unmarshal session data")
	}
	sessionData.UrlPath = b.baseUrl + "/banners"
	var banner models.Banner
	banners, err := banner.Gets()
	if err != nil {
		return err
	}
	data := fiber.Map{
		"sessionData": sessionData,
		"banners":     banners,
	}
	return c.Render("banner", data, "layouts/admin")
}

func (b *BannerController) CreateBanner(c *fiber.Ctx) error {
	session, err := b.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqBanner := new(models.Banner)
	err = c.BodyParser(reqBanner)
	if err != nil {
		return err
	}
	file, err := c.FormFile("image")
	if err != nil {
		return err
	}
	if file != nil {
		// Save file to root directory images:
		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		c.SaveFile(file, fmt.Sprintf("./static/img/banners/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))))
		if err != nil {
			return err
		}
		reqBanner.Image = fileName
	}

	err = reqBanner.Create()
	if err != nil {
		return err
	}
	return c.Redirect("/dashboard/banners")
}

func (b *BannerController) UpdateBanner(c *fiber.Ctx) error {
	session, err := b.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqBanner := new(models.Banner)
	err = c.BodyParser(reqBanner)
	if err != nil {
		return err
	}
	banner, err := reqBanner.Get(reqBanner.ID)
	if err != nil {
		return err
	}

	file, err := c.FormFile("image")
	if file != nil {
		//revome old file image
		oldImageName := banner.Image
		err = os.Remove(filepath.Join("static/img/banners/", oldImageName))
		if err != nil {
			return err
		}
		// Save file to root directory images:
		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		c.SaveFile(file, fmt.Sprintf("./static/img/banners/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))))
		if err != nil {
			return err
		}
		banner.Image = fileName
	}
	banner.Name = reqBanner.Name
	err = banner.Update()
	if err != nil {
		return err
	}

	return c.Redirect("/dashboard/banners")
}

func (b *BannerController) DeleteBanner(c *fiber.Ctx) error {
	session, err := b.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqBannerDelete := new(models.Banner)
	err = c.BodyParser(reqBannerDelete)
	if err != nil {
		return err
	}
	banners, err := reqBannerDelete.Gets()
	if err != nil {
		return err
	}

	var deletedBanner models.Banner
	for i, banner := range banners {
		if banner.ID == reqBannerDelete.ID {
			deletedBanner = banners[i]
			banners = append(banners[:i], banners[i+1:]...)
			break
		}
	}

	if deletedBanner.ID != 0 {
		imagePath := fmt.Sprintf("static/img/banners/%s", deletedBanner.Image)
		err = os.Remove(imagePath)
		if err != nil {
			fmt.Println("Error deleting image:", err)
		} else {
			fmt.Println("Image deleted:", imagePath)
		}
	}

	err = deletedBanner.Delete()
	if err != nil {
		return err
	}

	return c.Redirect("/dashboard/banners")
}
