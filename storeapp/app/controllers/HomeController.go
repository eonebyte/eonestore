package controllers

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type HomeController struct {
	baseUrl string
	session *session.Store
	setting models.Setting
}

func NewHomeController(session *session.Store, b string) *HomeController {
	return &HomeController{
		session: session,
		baseUrl: b,
	}
}

func (h *HomeController) Home(c *fiber.Ctx) error {
	session, err := h.session.Get(c)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = config.DB.Last(&setting).Error
	if err != nil {
		return err
	}
	

	pageStr := c.Query("page")
	categoryInt := c.QueryInt("category")
	searchQuery := c.Query("search")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	const itemsPerPage = 8

	var banner models.Banner
	banners, err := banner.Gets()
	if err != nil {
		return err
	}
	var categories []models.Category
	config.DB.Find(&categories)
	var items []models.Item
	config.DB.Find(&items)

	// Filter products based on category
	var filteredItems []models.Item
	if searchQuery != "" {
		for _, item := range items {
			if searchQuery == "" || strings.Contains(strings.ToLower(item.Name), strings.ToLower(searchQuery)) {
				filteredItems = append(filteredItems, item)
			}
		}
	} else if categoryInt != 0 {
		for _, item := range items {
			if categoryInt == 0 || item.CategoryID == categoryInt {
				filteredItems = append(filteredItems, item)
				
			}
		}
	} else {
		filteredItems = items
	}

	for i, _ := range filteredItems {
		filteredItems[i].DiscountPercentage = int(math.Ceil(calculateDiscountPercentage(filteredItems[i].OriginalPrice, filteredItems[i].FinalPrice)))
		filteredItems[i].OriginalPriceString = formatRupiahWithoutDecimal(float64(filteredItems[i].OriginalPrice))
		filteredItems[i].FinalPriceString = formatRupiahWithoutDecimal(float64(filteredItems[i].FinalPrice))
	}

	totalPages := int(math.Ceil(float64(len(filteredItems)) / float64(itemsPerPage)))
	startIdx := (page - 1) * itemsPerPage
	endIdx := int(math.Min(float64(startIdx+itemsPerPage), float64(len(filteredItems))))

	pagedItems := filteredItems[startIdx:endIdx]

	//get Session Data before Decode
	sessionDataJSON := session.Get("sessionData")
	if sessionDataJSON == nil {
		sessionData := SessionData{
			UserData:   models.User{},
			IsLoggedIn: "",
			Url:        "",
			Setting:    setting,
		}
		// Convert the User struct to JSON bytes
		NewSessionDataJSON, errMarshal := json.Marshal(sessionData)
		if errMarshal != nil {
			return c.SendString("Error marshaling user data to JSON.")
		}
		session.Set("sessionData", NewSessionDataJSON)
		if errSave := session.Save(); err != nil {
			panic(errSave)
		}

		sessionDataJSON = NewSessionDataJSON
		
	}

	// Convert the JSON data back to the User struct
	var sessionDataDecode SessionData
	err = json.Unmarshal(sessionDataJSON.([]byte), &sessionDataDecode)
	if err != nil {
		return c.SendString("Failed to unmarshal session data")
	}

	

	return c.Render("home", fiber.Map{
		"sessionData": sessionDataDecode,
		"categories":   categories,
		"banners":      banners,
		"items":        pagedItems,
		"page":         page,
		"nextPage":     page + 1,
		"previousPage": page - 1,
		"totalPages":   totalPages,
	}, "layouts/main")
}

func (h *HomeController) ProdukDetail(c *fiber.Ctx) error {
	session, err := h.session.Get(c)
	if err != nil {
		return err
	}
	var setting models.Setting
	err = config.DB.Last(&setting).Error
	if err != nil {
		return err
	}
	//get Session Data before Decode
	sessionDataJSON := session.Get("sessionData")
	if sessionDataJSON == nil {
		sessionData := SessionData{
			UserData:   models.User{},
			IsLoggedIn: "",
			Url:        "",
			Setting:    setting,
		}
		// Convert the User struct to JSON bytes
		NewSessionDataJSON, errMarshal := json.Marshal(sessionData)
		if errMarshal != nil {
			return c.SendString("Error marshaling user data to JSON.")
		}
		session.Set("sessionData", NewSessionDataJSON)
		if errSave := session.Save(); err != nil {
			panic(errSave)
		}

		sessionDataJSON = NewSessionDataJSON
		
	}

	// Convert the JSON data back to the User struct
	var sessionDataDecode SessionData
	err = json.Unmarshal(sessionDataJSON.([]byte), &sessionDataDecode)
	if err != nil {
		return c.SendString("Failed to unmarshal session data")
	}
	id := c.Params("id")
	var item models.Item
	config.DB.Preload("Category").Find(&item, id)

	var items []models.Item
	config.DB.Find(&items)
	var category models.Category
	config.DB.Find(&category, "id = ?", item.CategoryID)

	// Filter products based on category
	var filteredItems []models.Item
	if id != "" {
		for _, itm := range items {
			if itm.CategoryID == category.ID {
				filteredItems = append(filteredItems, itm)
			}
		}
	}

	const page = 1
	const itemsPerPage = 3

	startIdx := (page - 1) * itemsPerPage
	endIdx := int(math.Min(float64(startIdx+itemsPerPage), float64(len(filteredItems))))

	pagedItems := filteredItems[startIdx:endIdx]

	item.DiscountPercentage = int(math.Ceil(calculateDiscountPercentage(item.OriginalPrice, item.FinalPrice)))
	item.OriginalPriceString = formatRupiahWithoutDecimal(float64(item.OriginalPrice))
	item.FinalPriceString = formatRupiahWithoutDecimal(float64(item.FinalPrice))

	return c.Render("produk_detail", fiber.Map{
		"sessionData": sessionDataDecode,
		"item":        item,
		"items": pagedItems,
	}, "layouts/main")
}

func calculateDiscountPercentage(originalPrice, discountedPrice int) float64 {
	if originalPrice <= 0 {
		return 0.0
	}

	discount := float64(originalPrice - discountedPrice)
	discountPercentage := (discount / float64(originalPrice)) * 100
	return discountPercentage
}