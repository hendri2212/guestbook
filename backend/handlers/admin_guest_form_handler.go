package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"guestbook/backend/middleware"
	"guestbook/backend/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var guestFormSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type AdminGuestFormHandler struct {
	db *gorm.DB
}

type guestFormRequest struct {
	Name             *string         `json:"name"`
	PublicSlug       *string         `json:"public_slug"`
	Title            *string         `json:"title"`
	Description      *string         `json:"description"`
	IsActive         *bool           `json:"is_active"`
	RequirePhoto     *bool           `json:"require_photo"`
	RequireSignature *bool           `json:"require_signature"`
	Fields           json.RawMessage `json:"fields"`
}

type guestFormResponse struct {
	ID               string         `json:"id"`
	CompanyID        string         `json:"company_id"`
	Name             string         `json:"name"`
	PublicSlug       string         `json:"public_slug"`
	Title            string         `json:"title"`
	Description      *string        `json:"description"`
	IsActive         bool           `json:"is_active"`
	RequirePhoto     bool           `json:"require_photo"`
	RequireSignature bool           `json:"require_signature"`
	Fields           datatypes.JSON `json:"fields"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func NewAdminGuestFormHandler(db *gorm.DB) AdminGuestFormHandler {
	return AdminGuestFormHandler{db: db}
}

func (handler AdminGuestFormHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	var guestForms []models.GuestForm
	if err := handler.db.
		Where("company_id = ?", claims.CompanyID).
		Order("created_at DESC").
		Find(&guestForms).
		Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load guest forms")
		return
	}

	responses := make([]guestFormResponse, 0, len(guestForms))
	for _, guestForm := range guestForms {
		responses = append(responses, newGuestFormResponse(guestForm))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler AdminGuestFormHandler) Detail(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestForm, ok := handler.findGuestForm(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, newGuestFormResponse(guestForm))
}

func (handler AdminGuestFormHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	var request guestFormRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateCreate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fields, err := normalizeGuestFormFields(request.Fields)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if handler.publicSlugExists(*request.PublicSlug, "") {
		writeError(w, http.StatusConflict, "public_slug already exists")
		return
	}

	guestForm := models.GuestForm{
		CompanyID:        claims.CompanyID,
		Name:             valueOrDefault(request.Name, "Default Form"),
		PublicSlug:       *request.PublicSlug,
		Title:            *request.Title,
		Description:      trimOptionalString(request.Description),
		IsActive:         boolOrDefault(request.IsActive, true),
		RequirePhoto:     boolOrDefault(request.RequirePhoto, false),
		RequireSignature: boolOrDefault(request.RequireSignature, false),
		Fields:           fields,
	}

	if err := handler.db.Create(&guestForm).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create guest form")
		return
	}

	writeJSON(w, http.StatusCreated, newGuestFormResponse(guestForm))
}

func (handler AdminGuestFormHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestForm, ok := handler.findGuestForm(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	var request guestFormRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateUpdate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.PublicSlug != nil && handler.publicSlugExists(*request.PublicSlug, guestForm.ID) {
		writeError(w, http.StatusConflict, "public_slug already exists")
		return
	}

	if request.Name != nil {
		guestForm.Name = *request.Name
	}
	if request.PublicSlug != nil {
		guestForm.PublicSlug = *request.PublicSlug
	}
	if request.Title != nil {
		guestForm.Title = *request.Title
	}
	if request.Description != nil {
		guestForm.Description = trimOptionalString(request.Description)
	}
	if request.IsActive != nil {
		guestForm.IsActive = *request.IsActive
	}
	if request.RequirePhoto != nil {
		guestForm.RequirePhoto = *request.RequirePhoto
	}
	if request.RequireSignature != nil {
		guestForm.RequireSignature = *request.RequireSignature
	}
	if len(request.Fields) > 0 {
		fields, err := normalizeGuestFormFields(request.Fields)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		guestForm.Fields = fields
	}

	if err := handler.db.Save(&guestForm).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update guest form")
		return
	}

	writeJSON(w, http.StatusOK, newGuestFormResponse(guestForm))
}

func (handler AdminGuestFormHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestForm, ok := handler.findGuestForm(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	if err := handler.db.Delete(&guestForm).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete guest form")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler AdminGuestFormHandler) findGuestForm(w http.ResponseWriter, companyID string, id string) (models.GuestForm, bool) {
	var guestForm models.GuestForm
	if err := handler.db.
		Where("id = ? AND company_id = ?", id, companyID).
		First(&guestForm).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "guest form not found")
			return models.GuestForm{}, false
		}

		writeError(w, http.StatusInternalServerError, "failed to load guest form")
		return models.GuestForm{}, false
	}

	return guestForm, true
}

func (handler AdminGuestFormHandler) publicSlugExists(publicSlug string, exceptID string) bool {
	query := handler.db.Unscoped().Model(&models.GuestForm{}).
		Where("public_slug = ?", publicSlug)

	if exceptID != "" {
		query = query.Where("id <> ?", exceptID)
	}

	var count int64
	return query.Count(&count).Error == nil && count > 0
}

func (request *guestFormRequest) trim() {
	request.Name = trimOptionalString(request.Name)
	request.PublicSlug = normalizeSlug(request.PublicSlug)
	request.Title = trimOptionalString(request.Title)
	request.Description = trimOptionalString(request.Description)
}

func (request guestFormRequest) validateCreate() error {
	if request.PublicSlug == nil {
		return errors.New("public_slug is required")
	}
	if request.Title == nil {
		return errors.New("title is required")
	}

	return request.validateShared()
}

func (request guestFormRequest) validateUpdate() error {
	return request.validateShared()
}

func (request guestFormRequest) validateShared() error {
	if request.PublicSlug != nil && !guestFormSlugPattern.MatchString(*request.PublicSlug) {
		return errors.New("public_slug must use lowercase letters, numbers, and hyphens")
	}

	return nil
}

func normalizeGuestFormFields(raw json.RawMessage) (datatypes.JSON, error) {
	if len(raw) == 0 {
		return datatypes.JSON([]byte("[]")), nil
	}

	var fields []map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, errors.New("fields must be a json array")
	}
	if fields == nil {
		return nil, errors.New("fields must be a json array")
	}

	normalized, err := json.Marshal(fields)
	if err != nil {
		return nil, errors.New("fields is invalid")
	}

	return datatypes.JSON(normalized), nil
}

func normalizeSlug(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" {
		return nil
	}

	return &normalized
}

func valueOrDefault(value *string, fallback string) string {
	if value == nil {
		return fallback
	}

	return *value
}

func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

func authClaims(w http.ResponseWriter, r *http.Request) (*middleware.AuthClaims, bool) {
	claims, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}

	return claims, true
}

func newGuestFormResponse(guestForm models.GuestForm) guestFormResponse {
	return guestFormResponse{
		ID:               guestForm.ID,
		CompanyID:        guestForm.CompanyID,
		Name:             guestForm.Name,
		PublicSlug:       guestForm.PublicSlug,
		Title:            guestForm.Title,
		Description:      guestForm.Description,
		IsActive:         guestForm.IsActive,
		RequirePhoto:     guestForm.RequirePhoto,
		RequireSignature: guestForm.RequireSignature,
		Fields:           guestForm.Fields,
		CreatedAt:        guestForm.CreatedAt,
		UpdatedAt:        guestForm.UpdatedAt,
	}
}
