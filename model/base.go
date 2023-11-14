package model

import (
	"github.com/golang-module/carbon/v2"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64           `gorm:"column:id;primary_key" json:"id"`
	CreatedAt carbon.DateTime `gorm:"column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at" json:"deleted_at"`
}
