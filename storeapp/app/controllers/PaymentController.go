package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gopkg.in/gomail.v2"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
)

type PaymentController struct {
	baseUrl string
	session *session.Store
}

func NewPaymentController(session *session.Store, b string) *PaymentController {
	return &PaymentController{
		session: session,
		baseUrl: b,
	}
}

type RequestTransaksi struct {
	CustomerId     int           `json:"customer_id"`
	Address        string        `json:"address"`
	CourierCode    string        `json:"courier_code"`
	CourierService string        `json:"courier_service"`
	CostOngkir     int64         `json:"cost_ongkir"`
	Items          []models.Item `json:"items"`
}

type CustomerDetail struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

type RequestResult struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionStatus string `json:"transaction_status"`
	OrderId           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
	PdfUrl            string `json:"pdf_url"`
	FinishRedirectUrl string `json:"finish_redirect_url"`
}

func (p *PaymentController) PaymentController(c *fiber.Ctx) error {
	session, err := p.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	var setting models.Setting
	errSetting := config.DB.Find(&setting, "id = ?", 1).Error
	if errSetting != nil {
		return errSetting
	}
	requestTransaksi := new(RequestTransaksi)
	err = c.BodyParser(requestTransaksi)
	if err != nil {
		return err
	}

	var order models.Order
	err = config.DB.Find(&order, "order_no = ?", 1).Error
	if err != nil {
		return err
	}
	var user models.User
	err = config.DB.Find(&user, order.CustomerId).Error
	if err != nil {
		return err
	}

	var orderitems []models.OrderItem
	err = config.DB.Where("order_id = ?", order.ID).Find(&orderitems).Error
	if err != nil {
		return err
	}
	var items []models.Item
	subTotal := 0
	for _, i := range orderitems {
		var item models.Item
		err := config.DB.Find(&item, i.ItemId).Error
		if err != nil {
			return err
		}
		item.Quantity = i.Quantity
		items = append(items, item)
	}

	var transaction models.Transaction
	err = config.DB.Find(&transaction, "order_id = ?", order.ID).Error
	if err != nil {
		return err
	}

	//if transaction.OrderId == order.ID && transaction.Status == "Pending" {
	//	return c.JSON(fiber.Map{
	//		"token": transaction.Token,
	//	})
	//}

	// 1. Initiate Snap client
	var s = snap.Client{}
	s.New(setting.KeyMidtrans, midtrans.Sandbox)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  order.OrderNo,
			GrossAmt: int64(subTotal) + requestTransaksi.CostOngkir,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
	}

	// 3. Request create Snap transaction to Midtrans
	snapResp, errSnap := s.CreateTransaction(req)
	if errSnap != nil {
		return errSnap
	}
	fmt.Println("Response :", snapResp)

	var transactionModel models.Transaction
	//transactionModel.OrderId = order.ID
	transactionModel.Token = snapResp.Token
	transactionModel.Amount = int64(subTotal) + requestTransaksi.CostOngkir
	transactionModel.Status = "Pending"
	transactionModel.CreatedAt = time.Now()
	err = config.DB.Create(&transactionModel).Error
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"snap":  snapResp,
		"token": snapResp.Token,
	})
}

func (p *PaymentController) Payment(c *fiber.Ctx) error {
	session, err := p.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	var setting models.Setting
	errSetting := config.DB.Last(&setting).Error
	if errSetting != nil {
		return errSetting
	}
	requestTransaksi := new(RequestTransaksi)
	err = c.BodyParser(requestTransaksi)
	if err != nil {
		return err
	}

	items := requestTransaksi.Items

	//Data User
	var user models.User
	config.DB.Where("id = ?", requestTransaksi.CustomerId).First(&user)

	// 1. Initiate Snap client
	var s = snap.Client{}
	s.New(setting.KeyMidtrans, midtrans.Sandbox)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  models.GenerateOrderID("ORD"),
			GrossAmt: int64(calculateTotal(items)) + requestTransaksi.CostOngkir,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
			Phone: user.Phone,
			ShipAddr: &midtrans.CustomerAddress{
				Address: requestTransaksi.Address,
			},
		},
	}

	// 3. Request create Snap transaction to Midtrans
	snapResp, errSnap := s.CreateTransaction(req)
	if errSnap != nil {
		return errSnap
	}
	// Ambil data pesanan dari permintaan (request)
	newOrder := models.Order{
		OrderNo:         req.TransactionDetails.OrderID,
		CustomerName:    user.Name,
		CustomerId:      user.Id,
		Email:           user.Email,
		TotalAmount:     int(req.TransactionDetails.GrossAmt),
		OrderStatus:     "pending",
		ShippingAddress: requestTransaksi.Address,
		ShippingStatus:  "unshipping",
		CostOngkir:      requestTransaksi.CostOngkir,
		Courier:         "Provider : " + requestTransaksi.CourierCode + " | Service : " + requestTransaksi.CourierService,
		Token:           snapResp.Token,
		OrderDate:       time.Now(),
	}

	// Mulai transaksi
	tx := config.DB.Begin()

	// Simpan data pesanan ke dalam database
	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback() // Batalkan transaksi jika terjadi kesalahan
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create order"})
	}

	// Simpan data order item ke dalam database beserta qty
	for _, item := range items {
		orderItem := models.OrderItem{
			OrderID:  newOrder.ID,
			ItemId:   item.ID,
			Quantity: item.Quantity,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback() // Batalkan transaksi jika terjadi kesalahan
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create order items"})
		}
	}

	// Commit transaksi jika semuanya berhasil
	tx.Commit()

	fmt.Println("Response :", snapResp)

	return c.JSON(fiber.Map{
		"snap":  snapResp,
		"token": snapResp.Token,
	})
}

type MidtransOrderStatusResponse struct {
	StatusCode        string `json:"status_code"`
	TransactionID     string `json:"transaction_id"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	StatusMessage     string `json:"status_message"`
}

type Invoice struct {
	OrderNo     string
	Date        string
	Customer    string
	Items       []*models.Item
	SubTotal    int
	CostOngkir  string
	GrandTotal  string
	OrderStatus string
	// ... tambahkan data lain yang diperlukan
}

func (p *PaymentController) Notifikasi(c *fiber.Ctx) error {
	var setting models.Setting
	errSetting := config.DB.Last(&setting).Error
	if errSetting != nil {
		return errSetting
	}
	session, err := p.session.Get(c)
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
	sessionData.UrlPath = p.baseUrl + "/notifikasi"
	orderId := c.Query("order_id")
	//statusCode := c.Query("status_code")
	transactionStatus := c.Query("transaction_status")

	baseURL := os.Getenv("MIDTRANS_BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/"+orderId+"/status", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	req.Header.Set("Authorization", setting.AuthorizationMidtrans)

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send request"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get order status"})
	}

	var orderStatus MidtransOrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderStatus); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode response"})
	}

	var order models.Order
	err = config.DB.Preload("Items").Find(&order, "order_no = ?", orderId).Error
	if err != nil {
		return err
	}
	if orderStatus.TransactionStatus == "settlement" || orderStatus.TransactionStatus == "capture" {
		err = config.DB.Table("orders").Where("id = ?", order.ID).Updates(map[string]interface{}{"order_status": "settlement", "shipping_status": "unshipping"}).Error
		if err != nil {
			return err
		}
	}
	var user models.User
	err = config.DB.Find(&user, order.CustomerId).Error
	if err != nil {
		return err
	}

	var orderitems []models.OrderItem
	err = config.DB.Find(&orderitems, "order_id = ?", order.ID).Error
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		var orderItem models.OrderItem
		// Query data kuantitas berdasarkan ID order dan ID item
		config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
		item.Quantity = orderItem.Quantity
		//item.Checked = false
	}

	subTotal := 0
	for _, item := range order.Items {
		item.SubTotal += formatRupiahWithoutDecimal(float64(item.Price * item.Quantity))
		subTotal += item.Price * item.Quantity
	}

	dateInvoice := order.OrderDate
	costOngkir := formatRupiahWithoutDecimal(float64(order.CostOngkir))
	grandTotal := formatRupiahWithoutDecimal(float64(order.TotalAmount))

	//======= create pdf ======
	customerEmail := order.Email

	wkhtmltopdf.SetPath("static/wkhtmltopdf/bin/wkhtmltopdf.exe")

	// Data invoice dari struktur
	invoice := Invoice{
		OrderNo:     order.OrderNo,
		Date:        "2023-08-15",
		Customer:    order.CustomerName,
		Items:       order.Items,
		SubTotal:    subTotal,
		CostOngkir:  costOngkir,
		GrandTotal:  grandTotal,
		OrderStatus: order.OrderStatus,
	}

	// Baca template HTML dari berkas
	htmlTemplate, err := template.ParseFiles("app/views/invoice.html")
	if err != nil {
		fmt.Println("Error reading template:", err)
		return err
	}

	// Render template dengan data
	var tplBuffer bytes.Buffer
	err = htmlTemplate.Execute(&tplBuffer, invoice)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return err
	}

	// Generate PDF dari HTML
	pdfg, _ := wkhtmltopdf.NewPDFGenerator()
	pdfg.AddPage(wkhtmltopdf.NewPageReader(&tplBuffer))

	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)

	err = pdfg.Create()
	if err != nil {
		fmt.Println("Error generating PDF:", err)
		return err
	}

	fileNameInvoice := "static/pdf/" + fmt.Sprintf("%s%s", order.OrderNo, "-invoice.pdf")

	err = pdfg.WriteFile(fileNameInvoice)
	if err != nil {
		fmt.Println("Error writing PDF to file:", err)
		return err
	}

	subject := fmt.Sprintf("Invoice for Order #%s", fileNameInvoice)
	if err = sendEmailWithPDF(customerEmail, subject, "Invoice details are attached.", fileNameInvoice); err != nil {
		fmt.Println("Error sending email:", err)
	} else {
		fmt.Println("Invoice email with PDF sent successfully!")
	}

	return c.Render("notifikasi", fiber.Map{
		"sessionData":        sessionData,
		"setting":            setting,
		"order":              order,
		"cost_ongkir":        costOngkir,
		"sub_total":          formatRupiahWithoutDecimal(float64(subTotal)),
		"grand_total":        grandTotal,
		"order_date":         dateInvoice.Format("02-Jan-2006"),
		"user":               user,
		"transaction_status": transactionStatus,
	}, "layouts/admin")
}

func (p *PaymentController) ApiNotifikasi(c *fiber.Ctx) error {
	orderId := c.Query("order_id")
	var setting models.Setting
	errSetting := config.DB.Last(&setting).Error
	if errSetting != nil {
		return errSetting
	}

	baseURL := os.Getenv("MIDTRANS_BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/"+orderId+"/status", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	req.Header.Set("Authorization", setting.AuthorizationMidtrans)

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send request"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get order status"})
	}

	var orderStatus MidtransOrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderStatus); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode response"})
	}

	var order models.Order
	err = config.DB.Preload("Items").Find(&order, "order_no = ?", orderId).Error
	if err != nil {
		return err
	}
	if orderStatus.TransactionStatus == "settlement" || orderStatus.TransactionStatus == "capture" {
		err = config.DB.Table("orders").Where("id = ?", order.ID).Updates(map[string]interface{}{"order_status": "settlement", "shipping_status": "unshipping"}).Error
		if err != nil {
			return err
		}
	}
	if orderStatus.TransactionStatus == "expire" {
		err = config.DB.Table("orders").Where("id = ?", order.ID).Updates(map[string]interface{}{"order_status": "expire"}).Error
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"token":    order.Token,
			"redirect": "/dashboard/orders",
		})
	}
	if orderStatus.TransactionStatus == "deny" || orderStatus.TransactionStatus == "cancel" || orderStatus.TransactionStatus == "failure" {
		err = config.DB.Table("orders").Where("id = ?", order.ID).Delete(&order).Error
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"redirect": "/dashboard/orders",
		})
	}
	var user models.User
	err = config.DB.Find(&user, order.CustomerId).Error
	if err != nil {
		return err
	}

	var orderitems []models.OrderItem
	err = config.DB.Find(&orderitems, "order_id = ?", order.ID).Error
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		var orderItem models.OrderItem
		// Query data kuantitas berdasarkan ID order dan ID item
		config.DB.Where("order_id = ? AND item_id = ?", order.ID, item.ID).First(&orderItem)
		item.Quantity = orderItem.Quantity
		//item.Checked = false
	}

	subTotal := 0
	for _, item := range order.Items {
		item.SubTotal += formatRupiahWithoutDecimal(float64(item.Price * item.Quantity))
		subTotal += item.Price * item.Quantity
	}

	return c.JSON(fiber.Map{
		"token":    order.Token,
		"redirect": "/dashboard/orders",
	})
}

// Function to format a numeric value as Rupiah without decimals
func formatRupiahWithoutDecimal(amount float64) string {
	// Convert the amount to an integer (remove decimals)
	roundedAmount := int64(amount)

	// Add thousands separators to the amount
	formattedAmount := strconv.FormatInt(roundedAmount, 10)
	sep := "."
	for i := len(formattedAmount) - 3; i > 0; i -= 3 {
		formattedAmount = formattedAmount[:i] + sep + formattedAmount[i:]
	}

	// Add currency symbol (Rp)
	return "Rp. " + formattedAmount
}

func sendEmailWithPDF(toEmail, subject, body, pdfFilePath string) error {
	// Konfigurasi pengirim email
	var setting models.Setting
	err := config.DB.Last(&setting).Error
	if err != nil {
		return err
	}
	d := gomail.NewDialer(setting.SmtpHost, setting.SmtpPort, setting.Email, setting.EmailPassword)

	// Membuat email baru
	m := gomail.NewMessage()
	m.SetHeader("From", setting.Email)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Melampirkan file PDF
	m.Attach(pdfFilePath)

	// Mengirim email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
