package models

import "time"

type BaseModel struct {
	Id        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
