package model

import "gorm.io/gorm"

type Topic struct {
	gorm.Model
	Title   string
	Content string
	Uid     int
	Type_   int `gorm:"column:type"`
	Top     int
}

func (Topic) TableName() string {
	return "db_topic"
}
