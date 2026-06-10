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
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	IsActive *bool   `json:"is_active"`
}

type companyResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
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

	query := handler.db.Model(&models.Company{})
	if claims.Role != "owner" {
		query = query.Where("id = ?", claims.CompanyID)
	}

	var companies []models.Company
	if err := query.Order("created_at DESC").Find(&companies).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load companies")
		return
	}

	responses := make([]companyResponse, 0, len(companies))
	for _, company := range companies {
		responses = append(responses, handler.newCompanyResponse(company))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler AdminCompanyHandler) Detail(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	companyID := r.PathValue("id")
	if claims.Role != "owner" && companyID != claims.CompanyID {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}

	company, ok := handler.findCompany(w, companyID)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, handler.newCompanyResponse(company))
}

func (handler AdminCompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}
	if !requireOwner(w, claims) {
		return
	}

	var request companyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateCreate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	company := models.Company{
		Name:     *request.Name,
		Email:    trimOptionalString(request.Email),
		Phone:    trimOptionalString(request.Phone),
		Address:  trimOptionalString(request.Address),
		IsActive: boolOrDefault(request.IsActive, true),
	}

	if err := handler.db.Create(&company).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create company")
		return
	}

	writeJSON(w, http.StatusCreated, handler.newCompanyResponse(company))
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
	if claims.Role != "owner" && companyID != claims.CompanyID {
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

	if request.Name != nil {
		company.Name = *request.Name
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
	if claims.Role != "owner" && companyID != claims.CompanyID {
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

func (handler AdminCompanyHandler) newCompanyResponse(company models.Company) companyResponse {
	return companyResponse{
		ID:        company.ID,
		Name:      company.Name,
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
	request.Email = trimOptionalString(request.Email)
	request.Phone = trimOptionalString(request.Phone)
	request.Address = trimOptionalString(request.Address)
}

func (request companyRequest) validateCreate() error {
	if request.Name == nil || *request.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

func (request companyRequest) validateUpdate() error {
	if request.Name != nil && *request.Name == "" {
		return errors.New("name is required")
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
