package model

import (
	"gorm.io/gorm"
	"time"
)

type CollectAndLike interface {
	SetUid(uid uint)
	SetTid(tid uint)
	GetDeletedAt() gorm.DeletedAt
	GetType() interface{}
}
type Collect struct {
	Tid       uint `gorm:"primaryKey;autoIncrement:false"`
	Uid       uint `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*Collect) TableName() string {
	return "db_topic_interact_collect"
}
func (c *Collect) SetUid(uid uint) {
	c.Uid = uid
}
func (c *Collect) SetTid(tid uint) {
	c.Tid = tid
}
func (c *Collect) GetDeletedAt() gorm.DeletedAt {
	return c.DeletedAt
}
func (c *Collect) GetType() interface{} {
	return Collect{}
}
