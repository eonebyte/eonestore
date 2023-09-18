package controllers

import (
	"encoding/json"
	"fmt"
	ro "github.com/Bhinneka/go-rajaongkir"
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type SessionData struct {
	UserData   models.User
	IsLoggedIn string
	Url        string
	UrlPath    string
	Setting    models.Setting
}

type AuthController struct {
	baseUrl string
	session *session.Store
}

func NewAuthController(session *session.Store, b string) *AuthController {
	return &AuthController{
		session: session,
		baseUrl: b,
	}
}

func (a *AuthController) Dashboard(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
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
	sessionData.UrlPath = a.baseUrl + "/dashboard"

	var orders []models.Order
	err = config.DB.Preload("Items").Preload("Shipping").Find(&orders, "customer_id = ?", sessionData.UserData.Id).Error
	if err != nil {
		return err
	}
	var orderitems []models.OrderItem
	for _, order := range orders {
		err = config.DB.Find(&orderitems, "order_id = ?", order.ID).Error
		if err != nil {
			return err
		}
	}

	for i, order := range orders {
		for _, item := range order.Items {
			var orderItem models.OrderItem
			// Query data kuantitas berdasarkan ID order dan ID item
			config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
			item.Quantity = orderItem.Quantity
			item.SubTotal = formatRupiahWithoutDecimal(float64(item.Price * item.Quantity))
			orders[i].SubTotal += item.Price * item.Quantity
			orders[i].SubTotalString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal))
			orders[i].CostOngkirString = formatRupiahWithoutDecimal(float64(orders[i].CostOngkir))
			orders[i].TotalAmountString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal + int(orders[i].CostOngkir)))
		}
	}

	data := fiber.Map{
		"sessionData": sessionData,
		"orders":      orders,
	}

	return c.Render("dashboard", data, "layouts/admin")
}

func (a *AuthController) DashboardAdmin(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
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
	sessionData.UrlPath = a.baseUrl + "/dashboard/admin"

	var orders []models.Order
	err = config.DB.Preload("Items").Preload("Shipping").Find(&orders).Error
	if err != nil {
		return err
	}
	var orderitems []models.OrderItem
	for _, order := range orders {
		err = config.DB.Find(&orderitems, "order_id = ?", order.ID).Error
		if err != nil {
			return err
		}
	}

	for i, order := range orders {
		for _, item := range order.Items {
			var orderItem models.OrderItem
			// Query data kuantitas berdasarkan ID order dan ID item
			config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
			item.Quantity = orderItem.Quantity
			item.SubTotal = formatRupiahWithoutDecimal(float64(item.Price * item.Quantity))
			orders[i].SubTotal += item.Price * item.Quantity
			orders[i].SubTotalString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal))
			orders[i].CostOngkirString = formatRupiahWithoutDecimal(float64(orders[i].CostOngkir))
			orders[i].TotalAmountString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal + int(orders[i].CostOngkir)))
		}
	}

	// Get count User
	var countUser int64
	if err := config.DB.Model(&models.User{}).
		Where("role = ?", "user").
		Count(&countUser).Error; err != nil {
		return err
	}
	// Get Count Total Revenue
	var totalRevenue float64
	if err := config.DB.Model(&models.Order{}).
		Select("SUM(total_amount)").
		Where("order_status = ?", "settlement").
		Scan(&totalRevenue).Error; err != nil {
		return err
	}
	// Convert Count Total Revenue to String format RP.
	totalRevenueString := formatRupiahWithoutDecimal(totalRevenue)
	var totalSales int
	if err := config.DB.Model(&models.Order{}).
		Select("SUM(order_items.quantity)").
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.order_status = ?", "settlement").
		Scan(&totalSales).Error; err != nil {
		return err
	}

	data := fiber.Map{
		"sessionData":  sessionData,
		"orders":       orders,
		"countUser":    countUser,
		"totalRevenue": totalRevenueString,
		"totalSales":   totalSales,
	}

	return c.Render("dashboard_admin", data, "layouts/admin")
}

func (a *AuthController) Orders(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
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
	sessionData.UrlPath = a.baseUrl + "/orders"

	var orders []models.Order
	err = config.DB.Preload("Items").Preload("Shipping").Find(&orders, "customer_id = ?", sessionData.UserData.Id).Error
	if err != nil {
		return err
	}
	var orderitems []models.OrderItem
	for _, order := range orders {
		err = config.DB.Find(&orderitems, "order_id = ?", order.ID).Error
		if err != nil {
			return err
		}
	}

	for i, order := range orders {
		for _, item := range order.Items {
			var orderItem models.OrderItem
			// Query data kuantitas berdasarkan ID order dan ID item
			config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
			item.Quantity = orderItem.Quantity
			item.SubTotal = formatRupiahWithoutDecimal(float64(item.Price * item.Quantity))
			orders[i].SubTotal += item.Price * item.Quantity
			orders[i].SubTotalString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal))
			orders[i].CostOngkirString = formatRupiahWithoutDecimal(float64(orders[i].CostOngkir))
			orders[i].TotalAmountString = formatRupiahWithoutDecimal(float64(orders[i].SubTotal + int(orders[i].CostOngkir)))
		}
	}
	grandTotalBelumBayar := 0
	lengthBelumBayar := 0
	grandTotalSudahBayar := 0
	lengthSudahBayar := 0
	lengthDikirim := 0
	lengthDikemas := 0
	for _, order := range orders {
		switch order.OrderStatus {
		case "pending":
			for _, item := range order.Items {
				grandTotalBelumBayar += item.Price * item.Quantity
				lengthBelumBayar = len(order.Items)
			}
		case "settlement":
			for _, item := range order.Items {
				grandTotalSudahBayar += item.Price * item.Quantity
				lengthSudahBayar = len(order.Items)
			}
		}
		switch order.ShippingStatus {
		case "shipping":
			lengthDikirim = len(order.Items)
			break
		case "unshipping":
			lengthDikemas = len(order.Items)
			break
		}

	}

	data := fiber.Map{
		"sessionData":             sessionData,
		"orders":                  orders,
		"grand_total_belum_bayar": grandTotalBelumBayar,
		"grand_total_sudah_bayar": grandTotalSudahBayar,
		"item_length_belum_bayar": lengthBelumBayar,
		"item_length_sudah_bayar": lengthSudahBayar,
		"lengthDikirim":           lengthDikirim,
		"lengthDikemas":           lengthDikemas,
	}

	return c.Render("orders", data, "layouts/admin")
}

func (a *AuthController) Keranjang(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
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

	sessionData.UrlPath = a.baseUrl + "/keranjang"

	data := fiber.Map{
		"sessionData": sessionData,
	}

	return c.Render("keranjang", data, "layouts/admin")
}

func (a *AuthController) CheckoutPage(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
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

	sessionData.UrlPath = a.baseUrl + "/checkout"

	var setting models.Setting
	err = config.DB.Last(&setting).Error
	if err != nil {
		return err
	}

	//Init Raja Ongkir
	raja := ro.New(setting.KeyRajaOngkir, 10*time.Second)
	p := ro.QueryRequest{ProvinceID: ""}
	provinsi := raja.GetProvince(p)
	if provinsi.Error != nil {
		fmt.Println(provinsi.Error.Error())
	}
	city := ro.QueryRequest{CityID: ""}
	kota := raja.GetCity(city)
	if kota.Error != nil {
		fmt.Println(kota.Error.Error())
	}

	provinces := provinsi
	cities := kota
	//End Init Raja Ongkir

	//==========================================================

	data := fiber.Map{
		"sessionData": sessionData,
		"provinces":   provinces,
		"cities":      cities,
	}

	return c.Render("checkout", data, "layouts/admin")
}

func (a *AuthController) AuthPage(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
	if err != nil {
		return err
	}
	message := c.Query("message")

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
	return c.Render("auth", fiber.Map{
		"sessionData": sessionDataDecode,
		"message":     message,
	}, "layouts/main")
}

type LicenceRespopnse struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Status    string `json:"status"`
	Used      int    `json:"used"`
	UId       string `json:"uid"`
	UpdatedAt string `json:"updated_at"`
}

func (a *AuthController) HandleLogin(c *fiber.Ctx) error {
	var setting models.Setting
	errSet := config.DB.Last(&setting).Error
	if errSet != nil {
		return errSet
	}
	username := c.FormValue("username")
	password := c.FormValue("password")

	users, err := models.GetAllUsers()
	if err != nil {
		return err
	}

	//validate licence key
	baseURLWebService := "https://eone.biz.id/web_service"

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURLWebService+"/api/licence/"+setting.LicenseKey, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send request"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get order status"})
	}

	var licence LicenceRespopnse
	if err = json.NewDecoder(resp.Body).Decode(&licence); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode response"})
	}

	//end validasi licence key

	if licence.Status == "valid" && licence.Used == 1 {
		//Lakukan validasi pengguna
		for _, user := range users {
			if user.Role == "admin" {
				if user.UId == licence.UId {
					if user.Username == username && comparePassword([]byte(user.Password), password) {
						session, errSession := a.session.Get(c)
						if errSession != nil {
							return errSession
						}
						session.Set("authenticated", true)
						session.Set("username", username)
						userData, errGetUser := models.GetOneUser(username)
						if errGetUser != nil {
							return errGetUser
						}
						var setting models.Setting
						config.DB.Last(&setting)

						sessionData := SessionData{
							UserData:   userData,
							IsLoggedIn: "true",
							Url:        a.baseUrl,
							Setting:    setting,
						}
						// Convert the User struct to JSON bytes
						sessionDataJSON, errMarshal := json.Marshal(sessionData)
						if errMarshal != nil {
							return c.SendString("Error marshaling user data to JSON.")
						}
						session.Set("sessionData", sessionDataJSON)
						if errSave := session.Save(); err != nil {
							panic(errSave)
						}
						return c.Redirect("/dashboard/admin")
					}
				} 
			}
			if user.Role == "user" {
				if user.Username == username && comparePassword([]byte(user.Password), password) {
					session, errSession := a.session.Get(c)
					if errSession != nil {
						return errSession
					}
					session.Set("authenticated", true)
					session.Set("username", username)
					userData, errGetUser := models.GetOneUser(username)
					if errGetUser != nil {
						return errGetUser
					}
					var setting models.Setting
					config.DB.Last(&setting)

					sessionData := SessionData{
						UserData:   userData,
						IsLoggedIn: "true",
						Url:        a.baseUrl,
						Setting:    setting,
					}
					// Convert the User struct to JSON bytes
					sessionDataJSON, errMarshal := json.Marshal(sessionData)
					if errMarshal != nil {
						return c.SendString("Error marshaling user data to JSON.")
					}
					session.Set("sessionData", sessionDataJSON)
					if errSave := session.Save(); err != nil {
						panic(errSave)
					}
					return c.Redirect("/dashboard/keranjang")
				}
			}

		}
	} else {
		message := "Please provide a valid license key"

		// Redirect and pass the data as a query parameter
		redirectURL := fmt.Sprintf("/auth?message=%s", message)

		return c.Redirect(redirectURL, fiber.StatusSeeOther)
	}

	message := "Password / Username salah"

	// Redirect and pass the data as a query parameter
	redirectURL := fmt.Sprintf("/auth?message=%s", message)

	return c.Redirect(redirectURL, fiber.StatusSeeOther)
}

func (a *AuthController) HandleRegister(c *fiber.Ctx) error {
	var err error
	newUser := new(models.User)
	err = c.BodyParser(newUser)
	if err != nil {
		return err
	}

	username := newUser.Username
	password := newUser.Password

	//periksa apakah pengguna sudah ada
	users, err := models.GetAllUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Username == username {
			return c.Redirect("/auth")
		}
	}

	//daftarkan pengguna baru dengan password yang di-hash
	hashedPassword := hashPassword(password)
	newUser.Password = string(hashedPassword)
	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = time.Now()

	err = newUser.CreateUser()
	if err != nil {
		return err
	}

	session, err := a.session.Get(c)
	if err != nil {
		return err
	}
	session.Set("authenticated", true)
	session.Set("username", username)

	return c.Redirect("/dashboard/checkout")
}

func (a *AuthController) Logout(c *fiber.Ctx) error {
	session, err := a.session.Get(c)
	if err != nil {
		return err
	}
	session.Destroy()
	return c.Redirect("/auth")
}

func hashPassword(password string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

func comparePassword(hashedPassword []byte, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(plainPassword))
	return err == nil
}

//func calculateTotalAmount(items models.Order) int {
//	total := 0
//	for _, item := range items.Items {
//		total += item.Price * item.Quantity
//	}
//	return total
//}
