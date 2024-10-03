package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	Email    string `gorm:"unique"`
	Role     string
	Avatar   string
}

func (User) TableName() string {
	return "db_user"
}
