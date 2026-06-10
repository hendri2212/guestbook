package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"guestbook/backend/middleware"
	"guestbook/backend/models"

	"gorm.io/gorm"
)

type AdminCompanyHandler struct {
	db *gorm.DB
}

type companyRequest struct {
	Name     *string `json:"name"`
	Slug     *string `json:"slug"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	IsActive *bool   `json:"is_active"`
}

type companyResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Slug      string       `json:"slug"`
	Email     *string      `json:"email"`
	Phone     *string      `json:"phone"`
	Address   *string      `json:"address"`
	IsActive  bool         `json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Stats     companyStats `json:"stats"`
}

type companyStats struct {
	AdminUsers  int64 `json:"admin_users"`
	GuestForms  int64 `json:"guest_forms"`
	GuestVisits int64 `json:"guest_visits"`
}

func NewAdminCompanyHandler(db *gorm.DB) AdminCompanyHandler {
	return AdminCompanyHandler{db: db}
}

func (handler AdminCompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	company, ok := handler.findCompany(w, claims.CompanyID)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, []companyResponse{
		handler.newCompanyResponse(company),
	})
}

func (handler AdminCompanyHandler) Detail(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	companyID := r.PathValue("id")
	if companyID != claims.CompanyID {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}

	company, ok := handler.findCompany(w, companyID)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, handler.newCompanyResponse(company))
}

func (handler AdminCompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}
	if !requireOwnerOrAdmin(w, claims) {
		return
	}

	companyID := r.PathValue("id")
	if companyID != claims.CompanyID {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}

	company, ok := handler.findCompany(w, companyID)
	if !ok {
		return
	}

	var request companyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateUpdate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.Slug != nil && handler.slugExists(*request.Slug, company.ID) {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}

	if request.Name != nil {
		company.Name = *request.Name
	}
	if request.Slug != nil {
		company.Slug = *request.Slug
	}
	if request.Email != nil {
		company.Email = trimOptionalString(request.Email)
	}
	if request.Phone != nil {
		company.Phone = trimOptionalString(request.Phone)
	}
	if request.Address != nil {
		company.Address = trimOptionalString(request.Address)
	}
	if request.IsActive != nil {
		company.IsActive = *request.IsActive
	}

	if err := handler.db.Save(&company).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update company")
		return
	}

	writeJSON(w, http.StatusOK, handler.newCompanyResponse(company))
}

func (handler AdminCompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}
	if !requireOwner(w, claims) {
		return
	}

	companyID := r.PathValue("id")
	if companyID != claims.CompanyID {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}

	company, ok := handler.findCompany(w, companyID)
	if !ok {
		return
	}

	if err := handler.db.Delete(&company).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete company")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler AdminCompanyHandler) findCompany(w http.ResponseWriter, companyID string) (models.Company, bool) {
	var company models.Company
	if err := handler.db.
		Where("id = ?", companyID).
		First(&company).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "company not found")
			return models.Company{}, false
		}

		writeError(w, http.StatusInternalServerError, "failed to load company")
		return models.Company{}, false
	}

	return company, true
}

func (handler AdminCompanyHandler) slugExists(slug string, exceptID string) bool {
	query := handler.db.Unscoped().Model(&models.Company{}).
		Where("slug = ?", slug)

	if exceptID != "" {
		query = query.Where("id <> ?", exceptID)
	}

	var count int64
	return query.Count(&count).Error == nil && count > 0
}

func (handler AdminCompanyHandler) newCompanyResponse(company models.Company) companyResponse {
	return companyResponse{
		ID:        company.ID,
		Name:      company.Name,
		Slug:      company.Slug,
		Email:     company.Email,
		Phone:     company.Phone,
		Address:   company.Address,
		IsActive:  company.IsActive,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
		Stats:     handler.companyStats(company.ID),
	}
}

func (handler AdminCompanyHandler) companyStats(companyID string) companyStats {
	var stats companyStats

	handler.db.Model(&models.AdminUser{}).Where("company_id = ?", companyID).Count(&stats.AdminUsers)
	handler.db.Model(&models.GuestForm{}).Where("company_id = ?", companyID).Count(&stats.GuestForms)
	handler.db.Model(&models.GuestVisit{}).Where("company_id = ?", companyID).Count(&stats.GuestVisits)

	return stats
}

func (request *companyRequest) trim() {
	request.Name = trimStringPointer(request.Name)
	request.Slug = normalizeSlugPointer(request.Slug)
	request.Email = trimOptionalString(request.Email)
	request.Phone = trimOptionalString(request.Phone)
	request.Address = trimOptionalString(request.Address)
}

func (request companyRequest) validateUpdate() error {
	if request.Name != nil && *request.Name == "" {
		return errors.New("name is required")
	}
	if request.Slug != nil && !guestFormSlugPattern.MatchString(*request.Slug) {
		return errors.New("slug must use lowercase letters, numbers, and hyphens")
	}

	return nil
}

func trimStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func normalizeSlugPointer(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	return &normalized
}

func requireOwner(w http.ResponseWriter, claims *middleware.AuthClaims) bool {
	if claims.Role != "owner" {
		writeError(w, http.StatusForbidden, "owner role is required")
		return false
	}

	return true
}

func requireOwnerOrAdmin(w http.ResponseWriter, claims *middleware.AuthClaims) bool {
	if claims.Role != "owner" && claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "owner or admin role is required")
		return false
	}

	return true
}
