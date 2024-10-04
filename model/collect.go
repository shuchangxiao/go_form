package model

import (
	"gorm.io/gorm"
	"time"
)

type Collect struct {
	tid       uint64 `gorm:"primaryKey;autoIncrement:false"`
	uid       uint64 `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Collect) TableName() string {
	return "db_topic_interact_collect"
}
