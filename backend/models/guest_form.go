package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GuestForm struct {
	ID               string         `json:"id" gorm:"type:char(36);primaryKey"`
	CompanyID        string         `json:"company_id" gorm:"type:char(36);not null;index"`
	Name             string         `json:"name" gorm:"type:varchar(120);not null;default:'Default Form'"`
	PublicSlug       string         `json:"public_slug" gorm:"type:varchar(100);not null;uniqueIndex"`
	Title            string         `json:"title" gorm:"type:varchar(160);not null"`
	Description      *string        `json:"description" gorm:"type:text"`
	IsActive         bool           `json:"is_active" gorm:"not null;default:true"`
	RequirePhoto     bool           `json:"require_photo" gorm:"not null;default:false"`
	RequireSignature bool           `json:"require_signature" gorm:"not null;default:false"`
	Fields           datatypes.JSON `json:"fields" gorm:"type:json;not null"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	Company     Company      `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	GuestVisits []GuestVisit `json:"-" gorm:"foreignKey:FormID;constraint:OnDelete:RESTRICT"`
}

func (GuestForm) TableName() string {
	return "guest_forms"
}

func (guestForm *GuestForm) BeforeCreate(tx *gorm.DB) error {
	if guestForm.ID == "" {
		guestForm.ID = uuid.NewString()
	}

	if len(guestForm.Fields) == 0 {
		guestForm.Fields = datatypes.JSON([]byte("[]"))
	}

	return nil
}
