package model

type Advertise struct {
	ID             int64  `gorm:"primaryKey,column:id" json:"id"`
	Title          string `gorm:"column:title" json:"title"`
	Url            string `gorm:"column:url" json:"url"`
	Sort           int    `gorm:"column:sort" json:"sort"`
	ValidityPeriod int64  `gorm:"column:validity_period" json:"validity_period"`
	CreatedAt      int64  `gorm:"column:created_at" json:"created_at"`
}

func (Advertise) TableName() string {
	return "advertise"
}
