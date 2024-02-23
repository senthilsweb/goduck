package model

import (
	"github.com/jinzhu/gorm"
)

// User model
type User struct {
	gorm.Model
	Email        string `gorm:"type:varchar(100);unique_index" json:"email"`
	FirstName    string `gorm:"size:100;not null"              json:"first_name"`
	LastName     string `gorm:"size:100;not null"              json:"last_name"`
	Password     string `gorm:"size:100;not null"              json:"password"`
	ProfileImage string `gorm:"size:255"                       json:"profile_image"`
}
