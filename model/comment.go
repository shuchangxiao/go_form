package model

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	Uid     uint
	Tid     uint
	Content string
	Quote   uint
}

func (Comment) TableName() string {
	return "db_topic_comment"
}
