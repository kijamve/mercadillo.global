package models

import (
	"time"

	"gorm.io/gorm"

	H "mercadillo-global/helpers"
)

type User struct {
	ID              string         `json:"id" gorm:"type:char(36);primaryKey"`
	Email           string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password        string         `json:"-" gorm:"type:varchar(255);not null"`
	KYCStatus       string         `json:"kyc_status" gorm:"type:enum('pending','approved','rejected');default:'pending'"`
	PlanSlug        string         `json:"plan_slug" gorm:"type:varchar(100);default:'free'"`
	Status          string         `json:"status" gorm:"type:enum('active','inactive','suspended');default:'active'"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at" gorm:"type:timestamp null"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relations
	Products   []Product   `json:"products" gorm:"foreignKey:UserID"`
	Warehouses []Warehouse `json:"warehouses" gorm:"foreignKey:UserID"`
}

// GORM Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(u.ID) {
		u.ID = H.NewUUID()
	}
	return nil
}
