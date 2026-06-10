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

var validAdminUserRoles = map[string]bool{
	"owner": true,
	"admin": true,
	"staff": true,
}

type AdminUserHandler struct {
	db *gorm.DB
}

type adminUserRequest struct {
	CompanyID *string `json:"company_id"`
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Role      *string `json:"role"`
	IsActive  *bool   `json:"is_active"`
}

type adminUserResponse struct {
	ID          string     `json:"id"`
	CompanyID   string     `json:"company_id"`
	Company     company    `json:"company"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewAdminUserHandler(db *gorm.DB) AdminUserHandler {
	return AdminUserHandler{db: db}
}

func (handler AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	query := handler.db.Preload("Company")
	if claims.Role != "owner" {
		query = query.Where("company_id = ?", claims.CompanyID)
	}

	companyID := strings.TrimSpace(r.URL.Query().Get("company_id"))
	if companyID != "" {
		if claims.Role != "owner" && companyID != claims.CompanyID {
			writeError(w, http.StatusForbidden, "cannot filter users from another company")
			return
		}
		query = query.Where("company_id = ?", companyID)
	}

	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role != "" {
		if !validAdminUserRoles[role] {
			writeError(w, http.StatusBadRequest, "role is invalid")
			return
		}
		query = query.Where("role = ?", role)
	}

	status := strings.TrimSpace(r.URL.Query().Get("is_active"))
	if status != "" {
		switch status {
		case "true", "1":
			query = query.Where("is_active = ?", true)
		case "false", "0":
			query = query.Where("is_active = ?", false)
		default:
			writeError(w, http.StatusBadRequest, "is_active must be true or false")
			return
		}
	}

	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("(name LIKE ? OR email LIKE ?)", like, like)
	}

	var adminUsers []models.AdminUser
	if err := query.Order("created_at DESC").Find(&adminUsers).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load users")
		return
	}

	responses := make([]adminUserResponse, 0, len(adminUsers))
	for _, adminUser := range adminUsers {
		responses = append(responses, newAdminUserResponse(adminUser))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler AdminUserHandler) Detail(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	adminUser, ok := handler.findAdminUser(w, claims, r.PathValue("id"))
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, newAdminUserResponse(adminUser))
}

func (handler AdminUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}
	if !requireOwner(w, claims) {
		return
	}

	var request adminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateCreate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !handler.companyExists(*request.CompanyID) {
		writeError(w, http.StatusBadRequest, "company_id is invalid")
		return
	}
	if handler.emailExists(*request.Email, "") {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	adminUser := models.AdminUser{
		CompanyID:    *request.CompanyID,
		Name:         *request.Name,
		Email:        *request.Email,
		PasswordHash: string(passwordHash),
		Role:         *request.Role,
		IsActive:     boolOrDefault(request.IsActive, true),
	}

	if err := handler.db.Create(&adminUser).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if err := handler.db.Preload("Company").First(&adminUser, "id = ?", adminUser.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load created user")
		return
	}

	writeJSON(w, http.StatusCreated, newAdminUserResponse(adminUser))
}

func (handler AdminUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	adminUser, ok := handler.findAdminUser(w, claims, r.PathValue("id"))
	if !ok {
		return
	}
	if !canUpdateExistingUser(w, claims, adminUser) {
		return
	}

	var request adminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateUpdate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.CompanyID != nil && !requireOwner(w, claims) {
		return
	}
	if request.CompanyID != nil && !handler.companyExists(*request.CompanyID) {
		writeError(w, http.StatusBadRequest, "company_id is invalid")
		return
	}
	if request.Role != nil && !requireOwner(w, claims) {
		return
	}
	if request.IsActive != nil && !requireOwner(w, claims) {
		return
	}

	if request.Email != nil && handler.emailExists(*request.Email, adminUser.ID) {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	if request.CompanyID != nil {
		adminUser.CompanyID = *request.CompanyID
	}
	if request.Name != nil {
		adminUser.Name = *request.Name
	}
	if request.Email != nil {
		adminUser.Email = *request.Email
	}
	if request.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		adminUser.PasswordHash = string(passwordHash)
	}
	if request.Role != nil {
		adminUser.Role = *request.Role
	}
	if request.IsActive != nil {
		adminUser.IsActive = *request.IsActive
	}

	if err := handler.db.Save(&adminUser).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	if err := handler.db.Preload("Company").First(&adminUser, "id = ?", adminUser.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load updated user")
		return
	}

	writeJSON(w, http.StatusOK, newAdminUserResponse(adminUser))
}

func (handler AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}
	if !requireOwner(w, claims) {
		return
	}

	adminUser, ok := handler.findAdminUser(w, claims, r.PathValue("id"))
	if !ok {
		return
	}
	if adminUser.ID == claims.AdminUserID {
		writeError(w, http.StatusBadRequest, "cannot delete current user")
		return
	}

	if err := handler.db.Delete(&adminUser).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler AdminUserHandler) findAdminUser(w http.ResponseWriter, claims *middleware.AuthClaims, id string) (models.AdminUser, bool) {
	query := handler.db.Preload("Company").Where("id = ?", id)
	if claims.Role != "owner" {
		query = query.Where("company_id = ?", claims.CompanyID)
	}

	var adminUser models.AdminUser
	if err := query.First(&adminUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return models.AdminUser{}, false
		}

		writeError(w, http.StatusInternalServerError, "failed to load user")
		return models.AdminUser{}, false
	}

	return adminUser, true
}

func (handler AdminUserHandler) emailExists(email string, exceptID string) bool {
	query := handler.db.Unscoped().Model(&models.AdminUser{}).
		Where("email = ?", email)

	if exceptID != "" {
		query = query.Where("id <> ?", exceptID)
	}

	var count int64
	return query.Count(&count).Error == nil && count > 0
}

func (handler AdminUserHandler) companyExists(companyID string) bool {
	var count int64
	return handler.db.Model(&models.Company{}).
		Where("id = ?", companyID).
		Count(&count).
		Error == nil && count > 0
}

func (request *adminUserRequest) trim() {
	request.CompanyID = trimStringPointer(request.CompanyID)
	request.Name = trimStringPointer(request.Name)
	request.Email = normalizeEmailPointer(request.Email)
	request.Password = trimStringPointer(request.Password)
	request.Role = normalizeRolePointer(request.Role)
}

func (request adminUserRequest) validateCreate() error {
	if request.CompanyID == nil || *request.CompanyID == "" {
		return errors.New("company_id is required")
	}
	if request.Name == nil || *request.Name == "" {
		return errors.New("name is required")
	}
	if request.Email == nil || *request.Email == "" {
		return errors.New("email is required")
	}
	if request.Password == nil || *request.Password == "" {
		return errors.New("password is required")
	}
	if request.Role == nil {
		return errors.New("role is required")
	}

	return request.validateShared()
}

func (request adminUserRequest) validateUpdate() error {
	if request.CompanyID != nil && *request.CompanyID == "" {
		return errors.New("company_id is required")
	}
	if request.Name != nil && *request.Name == "" {
		return errors.New("name is required")
	}
	if request.Email != nil && *request.Email == "" {
		return errors.New("email is required")
	}
	if request.Password != nil && *request.Password == "" {
		return errors.New("password is required")
	}

	return request.validateShared()
}

func (request adminUserRequest) validateShared() error {
	if request.Role != nil && !validAdminUserRoles[*request.Role] {
		return errors.New("role is invalid")
	}

	return nil
}

func normalizeEmailPointer(value *string) *string {
	if value == nil {
		return nil
	}

	email := strings.ToLower(strings.TrimSpace(*value))
	return &email
}

func normalizeRolePointer(value *string) *string {
	if value == nil {
		return nil
	}

	role := strings.ToLower(strings.TrimSpace(*value))
	return &role
}

func canUpdateExistingUser(w http.ResponseWriter, claims *middleware.AuthClaims, adminUser models.AdminUser) bool {
	if claims.Role == "owner" {
		return true
	}

	if adminUser.ID != claims.AdminUserID {
		writeError(w, http.StatusForbidden, "admin and staff roles can only update their own user")
		return false
	}

	return true
}

func newAdminUserResponse(adminUser models.AdminUser) adminUserResponse {
	return adminUserResponse{
		ID:          adminUser.ID,
		CompanyID:   adminUser.CompanyID,
		Company:     newCompany(adminUser.Company),
		Name:        adminUser.Name,
		Email:       adminUser.Email,
		Role:        adminUser.Role,
		IsActive:    adminUser.IsActive,
		LastLoginAt: adminUser.LastLoginAt,
		CreatedAt:   adminUser.CreatedAt,
		UpdatedAt:   adminUser.UpdatedAt,
	}
}
