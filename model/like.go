package model

import (
	"gorm.io/gorm"
	"time"
)

type Like struct {
	Tid       uint `gorm:"primaryKey;autoIncrement:false"`
	Uid       uint `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*Like) TableName() string {
	return "db_topic_interact_like"
}
func (c *Like) SetUid(uid uint) {
	c.Uid = uid
}
func (c *Like) SetTid(tid uint) {
	c.Tid = tid
}
func (c *Like) GetDeletedAt() gorm.DeletedAt {
	return c.DeletedAt
}
func (c *Like) GetType() interface{} {
	return Like{}
}
