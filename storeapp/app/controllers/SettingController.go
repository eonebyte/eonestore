package controllers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type SettingController struct {
	baseUrl string
	session *session.Store
}

func NewSettingController(session *session.Store, b string) *SettingController {
	return &SettingController{
		session: session,
		baseUrl: b,
	}
}

func (s *SettingController) PageSetting(c *fiber.Ctx) error {
	session, err := s.session.Get(c)
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
	sessionData.UrlPath = s.baseUrl + "/settings"

	var setting models.Setting
	err = config.DB.Last(&setting).Error
	if err != nil {
		return err
	}

	data := fiber.Map{
		"sessionData": sessionData,
		"setting":     setting,
	}
	return c.Render("setting", data, "layouts/admin")
}

func (s *SettingController) UpdateSetting(c *fiber.Ctx) error {
	session, err := s.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqSetting := new(models.Setting)
	err = c.BodyParser(reqSetting)
	if err != nil {
		return err
	}
	setting, err := reqSetting.Get(reqSetting.ID)
	if err != nil {
		return err
	}
	fileLogo, err := c.FormFile("logo")
	if fileLogo != nil {

		// Save file to root directory images:
		fileNameLogo := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileLogo.Filename))
		c.SaveFile(fileLogo, fmt.Sprintf("./static/img/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileLogo.Filename))))
		if err != nil {
			return c.Redirect("/dashboard/settings")
		}
		setting.Logo = fileNameLogo
	}

	fileFavicon, err := c.FormFile("favicon")
	
	if fileFavicon != nil {
		// Save file favicon to root directory images:
		fileNameFavicon := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileFavicon.Filename))
		c.SaveFile(fileFavicon, fmt.Sprintf("./static/img/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileFavicon.Filename))))
		if err != nil {
			return c.Redirect("/dashboard/settings")
		}
		setting.Favicon = fileNameFavicon
	}
	if reqSetting.AppName != "" {
		setting.AppName = reqSetting.AppName
	}
	if reqSetting.LicenseKey != "" {
		setting.LicenseKey = reqSetting.LicenseKey
	}
	if reqSetting.Contact != "" {
		setting.Contact = reqSetting.Contact
	}
	if reqSetting.Address != "" {
		setting.Address = reqSetting.Address
	}
	if reqSetting.Email != "" {
		setting.Email = reqSetting.Email
	}
	if reqSetting.SmtpHost != "" {
		setting.SmtpHost = reqSetting.SmtpHost
	}
	if reqSetting.SmtpPort != 0 {
		setting.SmtpPort = reqSetting.SmtpPort
	}
	if reqSetting.EmailPassword != "" {
		setting.EmailPassword = reqSetting.EmailPassword
	}
	if reqSetting.InstagramUrl != "" {
		setting.InstagramUrl = reqSetting.InstagramUrl
	}
	if reqSetting.FacebookUrl != "" {
		setting.FacebookUrl = reqSetting.FacebookUrl
	}
	if reqSetting.TwitterUrl != "" {
		setting.TwitterUrl = reqSetting.TwitterUrl
	}
	if reqSetting.WaUrl != "" {
		setting.WaUrl = reqSetting.WaUrl
	}
	if reqSetting.TiktokUrl != "" {
		setting.TiktokUrl = reqSetting.TiktokUrl
	}
	if reqSetting.ShopeeUrl != "" {
		setting.ShopeeUrl = reqSetting.ShopeeUrl
	}
	if reqSetting.TokpedUrl != "" {
		setting.TokpedUrl = reqSetting.TokpedUrl
	}
	if reqSetting.BukalapakUrl != "" {
		setting.BukalapakUrl = reqSetting.BukalapakUrl
	}
	if reqSetting.KeyMidtrans != "" {
		setting.KeyMidtrans = reqSetting.KeyMidtrans
	}
	if reqSetting.KeyRajaOngkir != "" {
		setting.KeyRajaOngkir = reqSetting.KeyRajaOngkir
	}
	if reqSetting.AuthorizationMidtrans != "" {
		setting.AuthorizationMidtrans = reqSetting.AuthorizationMidtrans
	}
	if reqSetting.FooterText1 != "" {
		setting.FooterText1 = reqSetting.FooterText1
	}
	if reqSetting.FooterText2 != "" {
		setting.FooterText2 = reqSetting.FooterText2
	}
	setting.Update()
	return c.Redirect("/dashboard/settings")
}
