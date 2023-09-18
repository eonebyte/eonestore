package controllers

import (
	"encoding/json"
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"strconv"
)

type ShippingController struct {
	baseUrl string
	session *session.Store
}

func NewShippingController(session *session.Store, b string) *ShippingController {
	return &ShippingController{
		session: session,
		baseUrl: b,
	}
}

func (s *ShippingController) PageShipping(c *fiber.Ctx) error {
	var newShipping models.Shipping
	shippings, err := newShipping.Gets()
	if err != nil {
		return err
	}
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
	sessionData.UrlPath = s.baseUrl + "/shippings"
	data := fiber.Map{
		"sessionData": sessionData,
		"shippings":   shippings,
	}
	return c.Render("shipping", data, "layouts/admin")
}

func (s *ShippingController) UpdateShipping(c *fiber.Ctx) error {
	session, err := s.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}

	id, _ := strconv.Atoi(c.FormValue("id"))
	userId := c.FormValue("user_id")
	shippingStatus := c.FormValue("status_shipping")
	reqShipping := new(models.Shipping)
	c.BodyParser(reqShipping)
	shipping_by_orderID, err := reqShipping.Get(id)
	if err != nil {
		return err
	}
	var user models.User
	err = config.DB.First(&user, "id = ?", userId).Error
	if err != nil {
		return err
	}
	var order models.Order
	err = config.DB.Find(&order, "id = ?", id).Error
	if err != nil {
		return err
	}
	var shipping models.Shipping
	err = config.DB.Find(&shipping, "id = ?", shipping_by_orderID.ID).Error
	if err != nil {
		return err
	}
	config.DB.Model(&order).Update("shipping_status", shippingStatus)
	config.DB.Model(&shipping).Update("status_shipping", shippingStatus)

	if userId != "" {
		if user.Role == "user" {
			return c.Redirect("/dashboard/orders")
		}
	}

	return c.Redirect("/dashboard/shippings")
}
