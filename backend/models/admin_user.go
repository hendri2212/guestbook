package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminUser struct {
	ID           string         `json:"id" gorm:"type:char(36);primaryKey"`
	CompanyID    string         `json:"company_id" gorm:"type:char(36);not null;index"`
	Name         string         `json:"name" gorm:"type:varchar(120);not null"`
	Email        string         `json:"email" gorm:"type:varchar(160);not null;uniqueIndex:idx_admin_users_email"`
	PasswordHash string         `json:"-" gorm:"type:text;not null"`
	Role         string         `json:"role" gorm:"type:enum('owner','admin','staff');not null;default:'staff'"`
	IsActive     bool           `json:"is_active" gorm:"not null;default:true"`
	LastLoginAt  *time.Time     `json:"last_login_at" gorm:"type:datetime(6)"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Company Company `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

func (adminUser *AdminUser) BeforeCreate(tx *gorm.DB) error {
	if adminUser.ID == "" {
		adminUser.ID = uuid.NewString()
	}

	return nil
}
