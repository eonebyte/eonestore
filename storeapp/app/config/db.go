package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// database/database.go

var DB *gorm.DB

type Installation struct {
    gorm.Model
    Installed string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    DBName   string
}

func (c *DatabaseConfig) GetDSN() string {
    return c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func InitDatabase(dsn string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB = db
}

