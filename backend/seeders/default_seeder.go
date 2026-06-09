package seeders

import (
	"time"

	"guestbook/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	defaultCompanySlug    = "instansi-demo"
	defaultGuestFormSlug  = "buku-tamu-instansi-demo"
	defaultAdminPassword  = "password"
	defaultOwnerEmail     = "owner@instansi-demo.test"
	defaultAdminEmail     = "admin@instansi-demo.test"
	defaultGuestVisitName = "Budi Santoso"
)

func SeedDefaultData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		company, err := seedCompany(tx)
		if err != nil {
			return err
		}

		if err := seedAdminUsers(tx, company.ID); err != nil {
			return err
		}

		guestForm, err := seedGuestForm(tx, company.ID)
		if err != nil {
			return err
		}

		return seedGuestVisit(tx, company.ID, guestForm.ID)
	})
}

func seedCompany(db *gorm.DB) (*models.Company, error) {
	email := "info@instansi-demo.test"
	phone := "081234567890"
	address := "Jl. Demo No. 1"

	company := models.Company{
		Name:     "Instansi Demo",
		Slug:     defaultCompanySlug,
		Email:    &email,
		Phone:    &phone,
		Address:  &address,
		IsActive: true,
	}

	if err := db.Where("slug = ?", company.Slug).FirstOrCreate(&company).Error; err != nil {
		return nil, err
	}

	return &company, nil
}

func seedAdminUsers(db *gorm.DB, companyID string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUsers := []models.AdminUser{
		{
			CompanyID:    companyID,
			Name:         "Owner Instansi Demo",
			Email:        defaultOwnerEmail,
			PasswordHash: string(passwordHash),
			Role:         "owner",
			IsActive:     true,
		},
		{
			CompanyID:    companyID,
			Name:         "Admin Instansi Demo",
			Email:        defaultAdminEmail,
			PasswordHash: string(passwordHash),
			Role:         "admin",
			IsActive:     true,
		},
	}

	for _, adminUser := range adminUsers {
		if err := db.
			Where("company_id = ? AND email = ?", adminUser.CompanyID, adminUser.Email).
			FirstOrCreate(&adminUser).
			Error; err != nil {
			return err
		}
	}

	return nil
}

func seedGuestForm(db *gorm.DB, companyID string) (*models.GuestForm, error) {
	description := "Form buku tamu default untuk Instansi Demo."
	fields := datatypes.JSON([]byte(`[
		{"name":"identity_number","label":"Nomor Identitas","type":"text","required":false},
		{"name":"department","label":"Departemen Tujuan","type":"text","required":false}
	]`))

	guestForm := models.GuestForm{
		CompanyID:        companyID,
		Name:             "Form Buku Tamu Default",
		PublicSlug:       defaultGuestFormSlug,
		Title:            "Buku Tamu Instansi Demo",
		Description:      &description,
		IsActive:         true,
		RequirePhoto:     false,
		RequireSignature: false,
		Fields:           fields,
	}

	if err := db.Where("public_slug = ?", guestForm.PublicSlug).FirstOrCreate(&guestForm).Error; err != nil {
		return nil, err
	}

	return &guestForm, nil
}

func seedGuestVisit(db *gorm.DB, companyID string, guestFormID string) error {
	guestEmail := "budi.santoso@example.test"
	guestPhone := "081298765432"
	guestCompany := "PT Contoh Sejahtera"
	personToMeet := "Ibu Sari"
	metadata := datatypes.JSON([]byte(`{
		"identity_number":"3171000000000001",
		"department":"Administrasi"
	}`))

	now := time.Now()
	visitDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	guestVisit := models.GuestVisit{
		CompanyID:    companyID,
		FormID:       guestFormID,
		GuestName:    defaultGuestVisitName,
		GuestEmail:   &guestEmail,
		GuestPhone:   &guestPhone,
		GuestCompany: &guestCompany,
		Purpose:      "Meeting administrasi",
		PersonToMeet: &personToMeet,
		VisitDate:    visitDate,
		Status:       "checked_in",
		Metadata:     metadata,
	}

	return db.
		Where("company_id = ? AND form_id = ? AND guest_name = ? AND guest_email = ?", companyID, guestFormID, guestVisit.GuestName, guestEmail).
		FirstOrCreate(&guestVisit).
		Error
}
