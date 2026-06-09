package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GuestVisit struct {
	ID           string         `json:"id" gorm:"type:char(36);primaryKey"`
	CompanyID    string         `json:"company_id" gorm:"type:char(36);not null;index;index:idx_guest_visits_company_date,priority:1"`
	FormID       string         `json:"form_id" gorm:"type:char(36);not null;index"`
	GuestName    string         `json:"guest_name" gorm:"type:varchar(140);not null"`
	GuestEmail   *string        `json:"guest_email" gorm:"type:varchar(160)"`
	GuestPhone   *string        `json:"guest_phone" gorm:"type:varchar(40)"`
	GuestCompany *string        `json:"guest_company" gorm:"type:varchar(160)"`
	Purpose      string         `json:"purpose" gorm:"type:text;not null"`
	PersonToMeet *string        `json:"person_to_meet" gorm:"type:varchar(140)"`
	VisitDate    time.Time      `json:"visit_date" gorm:"type:date;not null;index:idx_guest_visits_company_date,priority:2"`
	CheckInAt    time.Time      `json:"check_in_at" gorm:"type:datetime(6);not null;autoCreateTime:milli"`
	CheckOutAt   *time.Time     `json:"check_out_at" gorm:"type:datetime(6)"`
	Status       string         `json:"status" gorm:"type:enum('checked_in','checked_out','cancelled');not null;default:'checked_in';index"`
	PhotoURL     *string        `json:"photo_url" gorm:"type:text"`
	SignatureURL *string        `json:"signature_url" gorm:"type:text"`
	Metadata     datatypes.JSON `json:"metadata" gorm:"type:json;not null"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Company Company   `json:"-" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Form    GuestForm `json:"-" gorm:"foreignKey:FormID;constraint:OnDelete:RESTRICT"`
}

func (GuestVisit) TableName() string {
	return "guest_visits"
}

func (guestVisit *GuestVisit) BeforeCreate(tx *gorm.DB) error {
	if guestVisit.ID == "" {
		guestVisit.ID = uuid.NewString()
	}

	if guestVisit.VisitDate.IsZero() {
		now := time.Now()
		guestVisit.VisitDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	if len(guestVisit.Metadata) == 0 {
		guestVisit.Metadata = datatypes.JSON([]byte("{}"))
	}

	return nil
}
