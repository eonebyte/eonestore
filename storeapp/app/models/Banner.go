package models

import "github.com/eonebyte/myapp/app/config"

type Banner struct {
	ID    int    `json:"id" form:"id"`
	Name  string `json:"name" form:"name"`
	Image string `json:"image" form:"image"`
}

func (b *Banner) Gets() ([]Banner, error) {
	var banners []Banner
	err := config.DB.Find(&banners).Error
	if err != nil {
		return nil, err
	}
	return banners, nil
}

func (b *Banner) Get(id int) (Banner, error) {
	var banner Banner
	result := config.DB.Where("id = ?", id).First(&banner)
	return banner, result.Error
}

func (b *Banner) Delete() error {
	config.DB.Delete(b)
	return nil
}

func (b *Banner) Update() error {
	config.DB.Save(b)
	return nil
}

func (b *Banner) Create() error {
	err := config.DB.Create(b).Error
	if err != nil {
		return err
	}
	return nil
}
