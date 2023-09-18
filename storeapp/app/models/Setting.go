package models

import (
	"github.com/eonebyte/myapp/app/config"
)

type Setting struct {
	ID                    int    `json:"id" form:"id"`
	AppName               string `json:"app_name" form:"app_name"`
	LicenseKey            string `json:"license_key" form:"license_key"`
	Contact               string `json:"contact" form:"contact"`
	Address               string `json:"address" form:"address"`
	Email                 string `json:"email" form:"email"`
	SmtpHost              string `json:"smtp_host" form:"smtp_host"`
	SmtpPort              int    `json:"smtp_port" form:"smtp_port"`
	EmailPassword         string `json:"email_password" form:"email_password"`
	Logo                  string `json:"logo" form:"logo"`
	Favicon               string `json:"favicon" form:"favicon"`
	InstagramUrl          string `json:"instagram_url" form:"instagram_url"`
	FacebookUrl           string `json:"facebook_url" form:"facebook_url"`
	WaUrl                 string `json:"wa_url" form:"wa_url"`
	TwitterUrl            string `json:"twitter_url" form:"twitter_url"`
	TiktokUrl             string `json:"tiktok_url" form:"tiktok_url"`
	ShopeeUrl             string `json:"shopee_url" form:"shopee_url"`
	TokpedUrl             string `json:"tokped_url" form:"tokped_url"`
	BukalapakUrl          string `json:"bukalapak_url" form:"bukalapak_url"`
	KeyRajaOngkir         string `json:"key_raja_ongkir" form:"key_raja_ongkir"`
	KeyMidtrans           string `json:"key_midtrans" form:"key_midtrans"`
	AuthorizationMidtrans string `json:"authorization_midtrans" form:"authorization_midtrans"`
	FooterText1           string `json:"footer_text_1" form:"footer_text_1"`
	FooterText2           string `json:"footer_text_2" form:"footer_text_2"`
	Installed             string `json:"installed" form:"installed"`
}

func (s *Setting) Get(id int) (Setting, error) {
	var setting Setting
	result := config.DB.Where("id = ?", id).Find(&setting)
	return setting, result.Error
}

func (s *Setting) Update() error {
	config.DB.Save(s)
	return nil
}
