package controllers

import (
	"encoding/json"

	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type CategoryController struct {
	baseUrl string
	session *session.Store
}

func NewCategoryController(session *session.Store, b string) *CategoryController {
	return &CategoryController{
		session: session,
		baseUrl: b,
	}
}

func (k *CategoryController) PageCategory(c *fiber.Ctx) error {
	session, err := k.session.Get(c)
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
	sessionData.UrlPath = k.baseUrl + "/categories"

	var categoriesData models.Category
	categories, err := categoriesData.Gets()
	if err != nil {
		return err
	}
	for i, _ := range categories {
		i++
	}

	cs := c.Cookies("csrf_")
	
	data := fiber.Map{
		"csrfToken": cs,
		"sessionData": sessionData,
		"categories":  categories,
	}
	return c.Render("category", data, "layouts/admin")
}

func (k *CategoryController) CreateCategory(c *fiber.Ctx) error {
	
	session, err := k.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	category := new(models.Category)
	err = c.BodyParser(category)
	if err != nil {
		return err
	}
	err = category.Create()
	if err != nil {
		return err
	}

	return c.Redirect("/dashboard/categories")
}

func (k *CategoryController) UpdateCategory(c *fiber.Ctx) error {
	session, err := k.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqCategory := new(models.Category)
	err = c.BodyParser(reqCategory)
	if err != nil {
		return err
	}
	category, err := reqCategory.Get(reqCategory.ID)
	if err != nil {
		return err
	}
	category.Name = reqCategory.Name
	category.Update()

	return c.Redirect("/dashboard/categories")
}

func (k *CategoryController) DeleteCategory(c *fiber.Ctx) error {
	session, err := k.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqCategoryDelete := new(models.Category)
	err = c.BodyParser(reqCategoryDelete)
	if err != nil {
		return err
	}
	category, err := reqCategoryDelete.Get(reqCategoryDelete.ID)
	if err != nil {
		return err
	}
	category.Delete()

	return c.Redirect("/dashboard/categories")
}
