package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	ID        string         `json:"id" gorm:"type:char(36);primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(160);not null"`
	Slug      string         `json:"slug" gorm:"type:varchar(80);not null;uniqueIndex"`
	Email     *string        `json:"email" gorm:"type:varchar(160)"`
	Phone     *string        `json:"phone" gorm:"type:varchar(40)"`
	Address   *string        `json:"address" gorm:"type:text"`
	IsActive  bool           `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	AdminUsers  []AdminUser  `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	GuestForms  []GuestForm  `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	GuestVisits []GuestVisit `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (Company) TableName() string {
	return "companies"
}

func (company *Company) BeforeCreate(tx *gorm.DB) error {
	if company.ID == "" {
		company.ID = uuid.NewString()
	}

	return nil
}
