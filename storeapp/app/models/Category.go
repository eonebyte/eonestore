package models

import "github.com/eonebyte/myapp/app/config"

type Category struct {
	ID   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

func (k *Category) Gets() ([]Category, error) {
	var categories []Category
	err := config.DB.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (k *Category) Get(id int) (Category, error) {
	var category Category
	result := config.DB.Where("id = ?", id).First(&category)
	return category, result.Error
}

func (k *Category) Delete() error {
	config.DB.Delete(k)
	return nil
}

func (k *Category) Update() error {
	config.DB.Save(k)
	return nil
}

func (k *Category) Create() error {
	err := config.DB.Create(k).Error
	if err != nil {
		return err
	}
	return nil
}
