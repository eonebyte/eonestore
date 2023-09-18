package models

import (
	"github.com/eonebyte/myapp/app/config"
	"time"
)

type User struct {
	Id        int       `json:"id" form:"id" gorm:"primary_key"`
	Username  string    `json:"username" form:"username"`
	Email     string    `json:"email" form:"email"`
	Password  string    `json:"password" form:"password"`
	Role      string    `json:"role" form:"role"`
	Name      string    `json:"name" form:"name"`
	Phone     string    `json:"phone" form:"phone"`
	UId       string    `json:"uid"`
	Address string `json:"address" form:"address"`
	Image string `json:"image" form:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Gets() ([]User, error) {
	var users []User
	err := config.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) Get(id int) (User, error) {
	var user User
	result := config.DB.Where("id = ?", id).First(&user)
	return user, result.Error
}

func (u *User) Update() error {
	config.DB.Save(u)
	return nil
}

func (u *User) Delete() error {
	config.DB.Delete(u)
	return nil
}

func GetAllUsers() ([]User, error) {
	var users []User
	err := config.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) CreateUser() error {
	err := config.DB.Create(u).Error
	if err != nil {
		return err
	}
	return nil
}

func GetOneUser(username string) (User, error) {
	var user User
	result := config.DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

func GetOneUserById(id int) (User, error) {
	var user User
	result := config.DB.Where("id = ?", id).First(&user)
	return user, result.Error
}
