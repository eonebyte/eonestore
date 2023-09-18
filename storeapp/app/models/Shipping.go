package models

import (
	"github.com/eonebyte/myapp/app/config"
	"time"
)

type Shipping struct {
	ID             int       `json:"id"`
	OrderId        int       `json:"order_id"`
	OrderNo        string    `json:"order_no"`
	Destination    string    `json:"destination"`
	NoResi         string    `json:"no_resi"`
	StatusShipping string    `json:"status_shipping"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *Shipping) Gets() ([]Shipping, error) {
	var shippings []Shipping
	err := config.DB.Find(&shippings).Error
	if err != nil {
		return nil, err
	}
	return shippings, nil
}

func (s *Shipping) Get(id int) (Shipping, error) {
	var shipping Shipping
	result := config.DB.Where("order_id = ?", id).Find(&shipping)
	return shipping, result.Error
}

func (s *Shipping) Update() error {
	config.DB.Save(s)
	return nil
}
