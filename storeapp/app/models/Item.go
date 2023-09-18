package models

import "github.com/eonebyte/myapp/app/config"

type Item struct {
	ID            int      `json:"id" form:"id"`
	Name          string   `json:"name" form:"name"`
	Price         int      `json:"price" form:"price"`
	OriginalPrice int      `json:"original_price" form:"original_price"`
	FinalPrice    int      `json:"final_price" form:"final_price"`
	PriceString   string   `json:"price_string" gorm:"-"`
	OriginalPriceString   string   `json:"original_price_string" gorm:"-"`
	FinalPriceString   string   `json:"final_price_string" gorm:"-"`
	DiscountPercentage  int   `json:"discount_percentage" gorm:"-"`
	Image         string   `json:"image" form:"image"`
	Description   string   `json:"description"`
	Quantity      int      `json:"quantity" form:"quantity"`
	Stock         int      `json:"stock" form:"stock"`
	SubTotal      string   `json:"sub_total" gorm:"-"`
	CategoryID    int      `json:"category_id" form:"category_id"`
	Orders        []*Order `json:"orders" gorm:"many2many:order_items;"`
	Category      Category `json:"category"`
}

type Items struct {
	Items []Item `json:"items" form:"items"`
}

func (i *Item) Gets() ([]Item, error) {
	var items []Item
	err := config.DB.Preload("Category").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (i *Item) Create() error {
	err := config.DB.Create(i).Error
	if err != nil {
		return err
	}
	return nil
}

func (i *Item) Get(id int) (Item, error) {
	var item Item
	result := config.DB.Where("id = ?", id).First(&item)
	return item, result.Error
}

func (i *Item) Update() error {
	config.DB.Save(i)
	return nil
}

func (i *Item) Delete() error {
	config.DB.Delete(i)
	return nil
}
