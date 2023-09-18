package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	ro "github.com/Bhinneka/go-rajaongkir"
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
)

type DatabaseConfig struct {
	PortApp         string `json:"port_app"`
	BaseUrl         string `json:"base_url"`
	DatabaseName    string `json:"databaseName"`
	DatabasePort    string `jason:"databasePort"`
	Host            string `json:"host"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	StatusInstall   string `json:"status_install"`
	StatusInstallDB string `json:"status_install_db"`
}


func PageInstall(c *fiber.Ctx) error {
	message := c.Query("message")
	data := fiber.Map{"message": message}
	return c.Render("install", data, "layouts/main")
}

func HandleInstall(c *fiber.Ctx) error {
	var err error
	newUser := new(models.User)
	err = c.BodyParser(newUser)
	if err != nil {
		return err
	}

	licenceKey := c.FormValue("licence_key")

	newSetting := new(models.Setting)
	err = c.BodyParser(newSetting)
	if err != nil {
		return err
	}

	//validate licence key
	baseURLWebService := "https://eone.biz.id/web_service"

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURLWebService+"/api/licence/"+licenceKey, nil)
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

	if licence.Status == "valid" && licence.Used == 0 {

		username := newUser.Username
		password := newUser.Password

		//periksa apakah pengguna sudah ada
		users, err := models.GetAllUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			if user.Username == username {
				return c.Redirect("/install")
			}
		}

		//daftarkan pengguna baru dengan password yang di-hash
		hashedPassword := hashPassword(password)
		newUser.Password = string(hashedPassword)
		newUser.CreatedAt = time.Now()
		newUser.UpdatedAt = time.Now()
		newUser.UId = fmt.Sprintf("%d", time.Now().UnixNano())

		err = newUser.CreateUser()
		if err != nil {
			return err
		}

		newSetting.LicenseKey = licenceKey
		err = config.DB.Create(&newSetting).Error
		if err != nil {
			return err
		}

		licenseData := LicenceRespopnse{
			Used: 1,
			Key:  licence.Key,
			UId:  newUser.UId,
		}

		// Mengkonversi struct menjadi JSON
		postData, errPost := json.Marshal(licenseData)
		if errPost != nil {
			return errPost
		}
		reqPost, errReq := http.NewRequest("PUT", baseURLWebService+"/api/licence/"+licenceKey, bytes.NewBuffer(postData))
		if errReq != nil {
			return errReq
		}

		reqPost.Header.Set("Content-Type", "application/json")

		resPost, errRes := client.Do(reqPost)
		if errRes != nil {
			return errRes
		}
		defer resPost.Body.Close()
		// created installed
		//update setting
		var updateSetting models.Setting
		err = config.DB.Model(&updateSetting).Where("id = ?", newSetting.ID).Update("installed", "true").Error
		if err != nil {
			return err
		}

		return c.Redirect("/auth")
	}

	message := "Please provide a valid license key"

	// Redirect and pass the data as a query parameter
	redirectURL := fmt.Sprintf("/install?message=%s", message)

	return c.Redirect(redirectURL, fiber.StatusSeeOther)
}

func GetCost(c *fiber.Ctx) error {
	var setting models.Setting
	errSetting := config.DB.Last(&setting).Error
	if errSetting != nil {
		return errSetting
	}
	destination := c.Params("destination")
	// Raja Ongkir Constructor, simply call ro.New
	// Parameter 1: "Raja Ongkir API KEY"
	// Parameter 2: HTTP Request Timeout
	raja := ro.New(setting.KeyRajaOngkir, 10*time.Second)
	//by default after construct, raja ongkir used Starter Env(Free account)
	//to change Env, simply just change the env
	//raja.Env = ro.Basic for basic account
	//or raja.Env = ro.Pro for pro account

	// Parameter 1: City Origin ID
	// Parameter 2: City Destination ID
	// Parameter 3: Item's Weight
	// Parameter 4: Courier's Name
	jne := ro.QueryRequest{Origin: "54", Destination: destination, Weight: 1000, Courier: "jne"}
	resultJne := raja.GetCost(jne)

	if resultJne.Error != nil {
		fmt.Println(resultJne.Error.Error())
	}

	costJne, ok := resultJne.Result.(ro.Cost)
	if !ok {
		fmt.Println("Result is not Cost")
	}
	tiki := ro.QueryRequest{Origin: "54", Destination: destination, Weight: 1000, Courier: "tiki"}
	resultTiki := raja.GetCost(tiki)

	if resultTiki.Error != nil {
		fmt.Println(resultTiki.Error.Error())
	}

	costTiki, ok := resultTiki.Result.(ro.Cost)
	if !ok {
		fmt.Println("Result is not Cost")
	}
	pos := ro.QueryRequest{Origin: "54", Destination: destination, Weight: 1000, Courier: "pos"}
	resultPos := raja.GetCost(pos)

	if resultPos.Error != nil {
		fmt.Println(resultPos.Error.Error())
	}

	costPos, ok := resultPos.Result.(ro.Cost)
	if !ok {
		fmt.Println("Result is not Cost")
	}

	courierService := "OKE"
	var courierServiceCost []interface{}

	for _, el := range costJne.Providers {
		for _, costs := range el.Costs {
			if costs.Service == courierService {
				courierServiceCost = append(courierServiceCost, costs)
			}
		}
	}

	data := fiber.Map{
		"JNE":          costJne,
		"JNE_PROVIDER": courierServiceCost,
		"TIKI":         costTiki,
		"POS":          costPos,
	}

	return c.JSON(data)
}

func GetProvider(c *fiber.Ctx) error {
	var setting models.Setting
	errSetting := config.DB.Last(&setting).Error
	if errSetting != nil {
		return errSetting
	}
	destination := c.Params("destination")
	courierName := c.Params("courier")
	pp, _ := url.QueryUnescape("provider")
	providerCourier := c.Query(pp)
	// Raja Ongkir Constructor, simply call ro.New
	// Parameter 1: "Raja Ongkir API KEY"
	// Parameter 2: HTTP Request Timeout
	raja := ro.New(setting.KeyRajaOngkir, 10*time.Second)
	//by default after construct, raja ongkir used Starter Env(Free account)
	//to change Env, simply just change the env
	//raja.Env = ro.Basic for basic account
	//or raja.Env = ro.Pro for pro account

	// Parameter 1: City Origin ID
	// Parameter 2: City Destination ID
	// Parameter 3: Item's Weight
	// Parameter 4: Courier's Name
	q := ro.QueryRequest{Origin: "54", Destination: destination, Weight: 1000, Courier: courierName}
	result := raja.GetCost(q)

	if result.Error != nil {
		fmt.Println(result.Error.Error())
	}

	cost, ok := result.Result.(ro.Cost)
	if !ok {
		fmt.Println("Result is not Cost")
	}

	courierService := providerCourier
	var courierProvider []interface{}

	for _, el := range cost.Providers {
		for _, costs := range el.Costs {
			if costs.Service == courierService {
				courierProvider = append(courierProvider, costs)
			}
		}
	}

	data := fiber.Map{
		"COURIER_PROVIDER": courierProvider,
	}
	fmt.Println(data)
	return c.JSON(data)
}

func (a *AuthController) MendapatkanItemBerdasarkanOrder(c *fiber.Ctx) error {
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
	var orders []models.Order
	err = config.DB.Preload("Items").Find(&orders, "customer_id = ?", sessionData.UserData.Id).Error
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

	for _, order := range orders {
		for _, item := range order.Items {
			var orderItem models.OrderItem
			// Query data kuantitas berdasarkan ID order dan ID item
			config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
			item.Quantity = orderItem.Quantity
		}
	}

	return c.JSON(orders)
}
