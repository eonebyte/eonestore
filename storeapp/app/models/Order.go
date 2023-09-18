package models

import (
	"github.com/eonebyte/myapp/app/config"
	"strconv"
	"sync"
	"time"
)

type Order struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	OrderNo           string    `json:"order_no"`
	CustomerName      string    `json:"customer_name"`
	CustomerId        int       `json:"customer_id"`
	Email             string    `json:"email"`
	TotalAmount       int       `json:"total_amount"`
	SubTotal          int       `json:"sub_total" gorm:"-"`
	SubTotalString    string    `json:"sub_total_string" gorm:"-"`
	TotalAmountString string    `json:"total_amount_string" gorm:"-"`
	OrderStatus       string    `json:"order_status"`
	OrderDate         time.Time `json:"order_date"`
	PaymentMethod     string    `json:"payment_method"`
	ShippingAddress   string    `json:"shipping_address"`
	Token             string    `json:"token"`
	CostOngkir        int64     `json:"cost_ongkir"`
	ShippingStatus    string    `json:"shipping_status"`
	Shipping          Shipping  `json:"shipping_id"`
	Courier           string    `json:"courier"`
	CostOngkirString  string    `json:"cost_ongkir_string" gorm:"-"`
	Items             []*Item   `json:"items" gorm:"many2many:order_items;"`
}

type OrderItem struct {
	ID       int `gorm:"primaryKey" json:"id"`
	OrderID  int `json:"order_id"`
	ItemId   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

var mutex = &sync.Mutex{}

func GenerateOrderID(orderPrefix string) string {
	var lastOrder Order
	_ = config.DB.Last(&lastOrder).Error

	mutex.Lock()
	defer mutex.Unlock()

	lastOrder.ID++
	currentTime := time.Now().Format("20060102150405") // Format time to "YYYYMMDDHHmmss"
	orderID := orderPrefix + currentTime + "-" + strconv.Itoa(lastOrder.ID)
	return orderID
}
