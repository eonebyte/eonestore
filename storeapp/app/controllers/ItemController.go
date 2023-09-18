package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type ItemController struct {
	baseUrl string
	session *session.Store
}

func NewItemController(session *session.Store, b string) *ItemController {
	return &ItemController{
		session: session,
		baseUrl: b,
	}
}

func (i *ItemController) ItemPage(c *fiber.Ctx) error {
	session, err := i.session.Get(c)
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
	sessionData.UrlPath = i.baseUrl + "/products"
	var item models.Item
	items, err := item.Gets()
	if err != nil {
		return err
	}
	for i, _ := range items {
		items[i].OriginalPriceString = formatRupiahWithoutDecimal(float64(items[i].OriginalPrice))
		items[i].FinalPriceString = formatRupiahWithoutDecimal(float64(items[i].FinalPrice))
	}
	var getCategories models.Category
	categories, err := getCategories.Gets()
	if err != nil {
		return err
	}
	data := fiber.Map{
		"sessionData": sessionData,
		"items":       items,
		"categories": categories,
	}
	return c.Render("product", data, "layouts/admin")
}

func (i *ItemController) CreateItem(c *fiber.Ctx) error {
	session, err := i.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqProduct := new(models.Item)
	err = c.BodyParser(reqProduct)
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
		c.SaveFile(file, fmt.Sprintf("./static/img/items/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))))
		if err != nil {
			return err
		}
		reqProduct.Image = fileName
	}

	err = reqProduct.Create()
	if err != nil {
		return err
	}
	return c.Redirect("/dashboard/products")
}

func (i *ItemController) UpdateItem(c *fiber.Ctx) error {
	session, err := i.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqItem := new(models.Item)
	err = c.BodyParser(reqItem)
	if err != nil {
		return err
	}

	categoryIDStr := c.FormValue("category_id")
		if categoryIDStr != "" {
			categoryID, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid category ID")
			}

			reqItem.CategoryID = categoryID
		}
	
	


	item, err := reqItem.Get(reqItem.ID)
	if err != nil {
		return err
	}

	file, err := c.FormFile("image")
	if file != nil {
		//revome old file image
		oldImageName := item.Image
		if oldImageName != "" {
			if _, err := os.Stat("./static/img/items/" + oldImageName); err == nil {
				err = os.Remove(filepath.Join("static/img/items/", oldImageName))
				if err != nil {
					return err
				}
			}  
				// Save file to root directory images:
				fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
				c.SaveFile(file, fmt.Sprintf("./static/img/items/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))))
				if err != nil {
					return err
				}
				item.Image = fileName
		}
	}
	if reqItem.Name != ""{
		item.Name = reqItem.Name
	}
	if reqItem.Price > 0 {
		item.Price = reqItem.Price
	}

	if reqItem.OriginalPrice > 0 {
		item.OriginalPrice = reqItem.OriginalPrice
	}

	if reqItem.FinalPrice > 0 {
		item.FinalPrice = reqItem.FinalPrice
	}
	
	if reqItem.CategoryID > 0 {
		item.CategoryID = reqItem.CategoryID
	}
	if reqItem.Stock != 0 {
		item.Stock = reqItem.Stock
	}
	if reqItem.Description != ""{
		item.Description = reqItem.Description
	}
	err = item.Update()
	if err != nil {
		return err
	}

	return c.Redirect("/dashboard/products")
}

func (i *ItemController) DeleteItem(c *fiber.Ctx) error {
	session, err := i.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqItemDelete := new(models.Item)
	err = c.BodyParser(reqItemDelete)
	if err != nil {
		return err
	}
	items, err := reqItemDelete.Gets()
	if err != nil {
		return err
	}

	var deletedItem models.Item
	for index, item := range items {
		if item.ID == reqItemDelete.ID {
			deletedItem = items[index]
			items = append(items[:index], items[index+1:]...)
			break
		}
	}

	if deletedItem.ID != 0 {
		imagePath := fmt.Sprintf("static/img/items/%s", deletedItem.Image)
		err = os.Remove(imagePath)
		if err != nil {
			fmt.Println("Error deleting image:", err)
		} else {
			fmt.Println("Image deleted:", imagePath)
		}
	}
	err = deletedItem.Delete()
	if err != nil {
		return err
	}

	return c.Redirect("/dashboard/products")
}
