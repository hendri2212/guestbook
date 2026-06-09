package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"guestbook/backend/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PublicHandler struct {
	db *gorm.DB
}

type submitGuestVisitRequest struct {
	GuestName    string          `json:"guest_name"`
	GuestEmail   *string         `json:"guest_email"`
	GuestPhone   string          `json:"guest_phone"`
	GuestCompany *string         `json:"guest_company"`
	Purpose      string          `json:"purpose"`
	PersonToMeet *string         `json:"person_to_meet"`
	PhotoURL     *string         `json:"photo_url"`
	SignatureURL *string         `json:"signature_url"`
	Metadata     json.RawMessage `json:"metadata"`
}

type submitGuestVisitResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CheckInAt time.Time `json:"check_in_at"`
}

type publicGuestFormResponse struct {
	ID          string         `json:"id"`
	PublicSlug  string         `json:"public_slug"`
	Title       string         `json:"title"`
	Description *string        `json:"description"`
	Fields      datatypes.JSON `json:"fields"`
	Company     string         `json:"company"`
}

func NewPublicHandler(db *gorm.DB) PublicHandler {
	return PublicHandler{db: db}
}

func (handler PublicHandler) GetGuestForm(w http.ResponseWriter, r *http.Request) {
	guestForm, ok := handler.findActiveGuestForm(w, r.PathValue("public_slug"))
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, publicGuestFormResponse{
		ID:          guestForm.ID,
		PublicSlug:  guestForm.PublicSlug,
		Title:       guestForm.Title,
		Description: guestForm.Description,
		Fields:      guestForm.Fields,
		Company:     guestForm.Company.Name,
	})
}

func (handler PublicHandler) SubmitGuestVisit(w http.ResponseWriter, r *http.Request) {
	var request submitGuestVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metadata, err := normalizeMetadata(request.Metadata)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	publicSlug := r.PathValue("public_slug")

	guestForm, ok := handler.findActiveGuestForm(w, publicSlug)
	if !ok {
		return
	}

	guestVisit := models.GuestVisit{
		CompanyID:    guestForm.CompanyID,
		FormID:       guestForm.ID,
		GuestName:    request.GuestName,
		GuestEmail:   request.GuestEmail,
		GuestPhone:   request.GuestPhone,
		GuestCompany: request.GuestCompany,
		Purpose:      request.Purpose,
		PersonToMeet: request.PersonToMeet,
		Status:       "checked_in",
		PhotoURL:     request.PhotoURL,
		SignatureURL: request.SignatureURL,
		Metadata:     metadata,
	}

	if err := handler.db.Create(&guestVisit).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save guest visit")
		return
	}

	writeJSON(w, http.StatusCreated, submitGuestVisitResponse{
		ID:        guestVisit.ID,
		Status:    guestVisit.Status,
		CheckInAt: guestVisit.CheckInAt,
	})
}

func (handler PublicHandler) findActiveGuestForm(w http.ResponseWriter, publicSlug string) (models.GuestForm, bool) {
	var guestForm models.GuestForm
	if err := handler.db.
		Preload("Company").
		Where("public_slug = ? AND is_active = ?", publicSlug, true).
		First(&guestForm).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "guest form not found")
			return models.GuestForm{}, false
		}

		writeError(w, http.StatusInternalServerError, "failed to load guest form")
		return models.GuestForm{}, false
	}

	if !guestForm.Company.IsActive {
		writeError(w, http.StatusNotFound, "guest form not found")
		return models.GuestForm{}, false
	}

	return guestForm, true
}

func (request *submitGuestVisitRequest) trim() {
	request.GuestName = strings.TrimSpace(request.GuestName)
	request.GuestPhone = strings.TrimSpace(request.GuestPhone)
	request.Purpose = strings.TrimSpace(request.Purpose)
	request.GuestEmail = trimOptionalString(request.GuestEmail)
	request.GuestCompany = trimOptionalString(request.GuestCompany)
	request.PersonToMeet = trimOptionalString(request.PersonToMeet)
	request.PhotoURL = trimOptionalString(request.PhotoURL)
	request.SignatureURL = trimOptionalString(request.SignatureURL)
}

func (request submitGuestVisitRequest) validate() error {
	if request.GuestName == "" {
		return errors.New("guest_name is required")
	}

	if request.GuestPhone == "" {
		return errors.New("guest_phone is required")
	}

	if request.Purpose == "" {
		return errors.New("purpose is required")
	}

	return nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeMetadata(raw json.RawMessage) (datatypes.JSON, error) {
	if len(raw) == 0 {
		return datatypes.JSON([]byte("{}")), nil
	}

	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil, errors.New("metadata must be a json object")
	}
	if metadata == nil {
		return nil, errors.New("metadata must be a json object")
	}

	normalized, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.New("metadata is invalid")
	}

	return datatypes.JSON(normalized), nil
}
