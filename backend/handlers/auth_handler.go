package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"guestbook/backend/middleware"
	"guestbook/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db              *gorm.DB
	authenticator   middleware.Authenticator
	tokenExpiration time.Duration
}

type loginRequest struct {
	CompanySlug string `json:"company_slug"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      authUser  `json:"user"`
	Company   company   `json:"company"`
}

type authUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type meResponse struct {
	User    authUser `json:"user"`
	Company company  `json:"company"`
}

func NewAuthHandler(db *gorm.DB, authenticator middleware.Authenticator, tokenExpiration time.Duration) AuthHandler {
	return AuthHandler{
		db:              db,
		authenticator:   authenticator,
		tokenExpiration: tokenExpiration,
	}
}

func (handler AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var companyRecord models.Company
	if err := handler.db.
		Where("slug = ? AND is_active = ?", request.CompanySlug, true).
		First(&companyRecord).
		Error; err != nil {
		writeInvalidCredentials(w, err)
		return
	}

	var adminUser models.AdminUser
	if err := handler.db.
		Where("company_id = ? AND email = ? AND is_active = ?", companyRecord.ID, request.Email, true).
		First(&adminUser).
		Error; err != nil {
		writeInvalidCredentials(w, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte(request.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	now := time.Now()
	if err := handler.db.Model(&adminUser).Update("last_login_at", now).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update last login")
		return
	}

	expiresAt := now.Add(handler.tokenExpiration)
	token, err := handler.authenticator.GenerateToken(adminUser.ID, adminUser.CompanyID, adminUser.Email, adminUser.Role, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      newAuthUser(adminUser),
		Company:   newCompany(companyRecord),
	})
}

func (handler AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var adminUser models.AdminUser
	if err := handler.db.
		Where("id = ? AND company_id = ? AND is_active = ?", claims.AdminUserID, claims.CompanyID, true).
		First(&adminUser).
		Error; err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var companyRecord models.Company
	if err := handler.db.
		Where("id = ? AND is_active = ?", claims.CompanyID, true).
		First(&companyRecord).
		Error; err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		User:    newAuthUser(adminUser),
		Company: newCompany(companyRecord),
	})
}

func (request *loginRequest) trim() {
	request.CompanySlug = strings.TrimSpace(request.CompanySlug)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Password = strings.TrimSpace(request.Password)
}

func (request loginRequest) validate() error {
	if request.CompanySlug == "" {
		return errors.New("company_slug is required")
	}

	if request.Email == "" {
		return errors.New("email is required")
	}

	if request.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func writeInvalidCredentials(w http.ResponseWriter, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	writeError(w, http.StatusInternalServerError, "failed to process login")
}

func newAuthUser(adminUser models.AdminUser) authUser {
	return authUser{
		ID:    adminUser.ID,
		Name:  adminUser.Name,
		Email: adminUser.Email,
		Role:  adminUser.Role,
	}
}

func newCompany(companyRecord models.Company) company {
	return company{
		ID:   companyRecord.ID,
		Name: companyRecord.Name,
		Slug: companyRecord.Slug,
	}
}
