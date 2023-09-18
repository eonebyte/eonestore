package controllers

import (
	"encoding/json"
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"time"
)

type OrderController struct {
	baseUrl string
	session *session.Store
}

func NewOrderController(session *session.Store, b string) *OrderController {
	return &OrderController{
		session: session,
		baseUrl: b,
	}
}

// Fungsi untuk menghitung total harga dari slice Item
func calculateTotal(items []models.Item) int {
	total := 0
	for _, item := range items {
		total += item.Price * item.Quantity
	}
	return total
}

func (o *OrderController) AdminOrders(c *fiber.Ctx) error {
	session, err := o.session.Get(c)
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
	sessionData.UrlPath = o.baseUrl + "/admin-orders"

	var orders []models.Order
	err = config.DB.Preload("Items").Find(&orders).Error
	if err != nil {
		return err
	}

	data := fiber.Map{
		"sessionData": sessionData,
		"orders":      orders,
	}

	return c.Render("admin_orders", data, "layouts/admin")
}

func (o *OrderController) DeleteOrder(c *fiber.Ctx) error {
	session, err := o.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	id := c.FormValue("id")
	var order models.Order
	err = config.DB.Find(&order, "id = ?", id).Error
	if err != nil {
		return err
	}
	config.DB.Delete(&order)

	return c.Redirect("/dashboard/admin-orders")
}

func (o *OrderController) AturPengiriman(c *fiber.Ctx) error {
	session, err := o.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	id := c.FormValue("id")
	no_resi := c.FormValue("resi_no")

	var order models.Order
	err = config.DB.Find(&order, "id = ?", id).Error
	if err != nil {
		return err
	}
	// begin a transaction
	tx := config.DB.Begin()

	newShipping := models.Shipping{
		OrderId:        order.ID,
		OrderNo:        order.OrderNo,
		Destination:    order.ShippingAddress,
		NoResi:         no_resi,
		StatusShipping: "Sedang dikirim",
		CreatedAt:      time.Now(),
	}
	err = tx.Create(&newShipping).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Model(&order).Update("shipping_status", "shipping")
	tx.Commit()
	return c.Redirect("/dashboard/shippings")
}
