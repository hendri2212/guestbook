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

var validGuestVisitStatuses = map[string]bool{
	"checked_in":  true,
	"checked_out": true,
	"cancelled":   true,
}

type AdminGuestVisitHandler struct {
	db *gorm.DB
}

type guestVisitRequest struct {
	FormID       *string         `json:"form_id"`
	GuestName    *string         `json:"guest_name"`
	GuestEmail   *string         `json:"guest_email"`
	GuestPhone   *string         `json:"guest_phone"`
	GuestCompany *string         `json:"guest_company"`
	Purpose      *string         `json:"purpose"`
	PersonToMeet *string         `json:"person_to_meet"`
	VisitDate    *string         `json:"visit_date"`
	CheckOutAt   *string         `json:"check_out_at"`
	Status       *string         `json:"status"`
	PhotoURL     *string         `json:"photo_url"`
	SignatureURL *string         `json:"signature_url"`
	Metadata     json.RawMessage `json:"metadata"`
}

type guestVisitResponse struct {
	ID           string         `json:"id"`
	CompanyID    string         `json:"company_id"`
	FormID       string         `json:"form_id"`
	Form         guestVisitForm `json:"form"`
	GuestName    string         `json:"guest_name"`
	GuestEmail   *string        `json:"guest_email"`
	GuestPhone   string         `json:"guest_phone"`
	GuestCompany *string        `json:"guest_company"`
	Purpose      string         `json:"purpose"`
	PersonToMeet *string        `json:"person_to_meet"`
	VisitDate    string         `json:"visit_date"`
	CheckInAt    time.Time      `json:"check_in_at"`
	CheckOutAt   *time.Time     `json:"check_out_at"`
	Status       string         `json:"status"`
	PhotoURL     *string        `json:"photo_url"`
	SignatureURL *string        `json:"signature_url"`
	Metadata     datatypes.JSON `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type guestVisitForm struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicSlug string `json:"public_slug"`
	Title      string `json:"title"`
}

func NewAdminGuestVisitHandler(db *gorm.DB) AdminGuestVisitHandler {
	return AdminGuestVisitHandler{db: db}
}

func (handler AdminGuestVisitHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	query := handler.db.
		Preload("Form").
		Where("company_id = ?", claims.CompanyID)

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" {
		if !validGuestVisitStatuses[status] {
			writeError(w, http.StatusBadRequest, "status is invalid")
			return
		}
		query = query.Where("status = ?", status)
	}

	formID := strings.TrimSpace(r.URL.Query().Get("form_id"))
	if formID != "" {
		query = query.Where("form_id = ?", formID)
	}

	dateFrom := strings.TrimSpace(r.URL.Query().Get("date_from"))
	if dateFrom != "" {
		parsed, err := parseDate(dateFrom)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date_from must use YYYY-MM-DD")
			return
		}
		query = query.Where("visit_date >= ?", parsed)
	}

	dateTo := strings.TrimSpace(r.URL.Query().Get("date_to"))
	if dateTo != "" {
		parsed, err := parseDate(dateTo)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date_to must use YYYY-MM-DD")
			return
		}
		query = query.Where("visit_date <= ?", parsed)
	}

	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"(guest_name LIKE ? OR guest_email LIKE ? OR guest_phone LIKE ? OR guest_company LIKE ?)",
			like,
			like,
			like,
			like,
		)
	}

	var guestVisits []models.GuestVisit
	if err := query.
		Order("check_in_at DESC").
		Find(&guestVisits).
		Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load guest visits")
		return
	}

	responses := make([]guestVisitResponse, 0, len(guestVisits))
	for _, guestVisit := range guestVisits {
		responses = append(responses, newGuestVisitResponse(guestVisit))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler AdminGuestVisitHandler) Detail(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestVisit, ok := handler.findGuestVisit(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, newGuestVisitResponse(guestVisit))
}

func (handler AdminGuestVisitHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	var request guestVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateCreate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !handler.guestFormExists(claims.CompanyID, *request.FormID) {
		writeError(w, http.StatusBadRequest, "form_id is invalid")
		return
	}

	metadata, err := normalizeMetadata(request.Metadata)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	visitDate, err := parseOptionalDate(request.VisitDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "visit_date must use YYYY-MM-DD")
		return
	}

	checkOutAt, err := parseOptionalDateTime(request.CheckOutAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "check_out_at must use RFC3339 format")
		return
	}

	status := valueOrDefault(request.Status, "checked_in")
	guestVisit := models.GuestVisit{
		CompanyID:    claims.CompanyID,
		FormID:       *request.FormID,
		GuestName:    *request.GuestName,
		GuestEmail:   trimOptionalString(request.GuestEmail),
		GuestPhone:   *request.GuestPhone,
		GuestCompany: trimOptionalString(request.GuestCompany),
		Purpose:      *request.Purpose,
		PersonToMeet: trimOptionalString(request.PersonToMeet),
		VisitDate:    visitDate,
		CheckOutAt:   checkOutAt,
		Status:       status,
		PhotoURL:     trimOptionalString(request.PhotoURL),
		SignatureURL: trimOptionalString(request.SignatureURL),
		Metadata:     metadata,
	}

	if err := handler.db.Create(&guestVisit).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create guest visit")
		return
	}

	if err := handler.db.Preload("Form").First(&guestVisit, "id = ?", guestVisit.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load created guest visit")
		return
	}

	writeJSON(w, http.StatusCreated, newGuestVisitResponse(guestVisit))
}

func (handler AdminGuestVisitHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestVisit, ok := handler.findGuestVisit(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	var request guestVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.trim()
	if err := request.validateUpdate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.FormID != nil {
		if !handler.guestFormExists(claims.CompanyID, *request.FormID) {
			writeError(w, http.StatusBadRequest, "form_id is invalid")
			return
		}
		guestVisit.FormID = *request.FormID
	}
	if request.GuestName != nil {
		guestVisit.GuestName = *request.GuestName
	}
	if request.GuestEmail != nil {
		guestVisit.GuestEmail = trimOptionalString(request.GuestEmail)
	}
	if request.GuestPhone != nil {
		guestVisit.GuestPhone = *request.GuestPhone
	}
	if request.GuestCompany != nil {
		guestVisit.GuestCompany = trimOptionalString(request.GuestCompany)
	}
	if request.Purpose != nil {
		guestVisit.Purpose = *request.Purpose
	}
	if request.PersonToMeet != nil {
		guestVisit.PersonToMeet = trimOptionalString(request.PersonToMeet)
	}
	if request.VisitDate != nil {
		visitDate, err := parseDate(*request.VisitDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "visit_date must use YYYY-MM-DD")
			return
		}
		guestVisit.VisitDate = visitDate
	}
	if request.CheckOutAt != nil {
		checkOutAt, err := parseOptionalDateTime(request.CheckOutAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "check_out_at must use RFC3339 format")
			return
		}
		guestVisit.CheckOutAt = checkOutAt
	}
	if request.Status != nil {
		guestVisit.Status = *request.Status
	}
	if request.PhotoURL != nil {
		guestVisit.PhotoURL = trimOptionalString(request.PhotoURL)
	}
	if request.SignatureURL != nil {
		guestVisit.SignatureURL = trimOptionalString(request.SignatureURL)
	}
	if len(request.Metadata) > 0 {
		metadata, err := normalizeMetadata(request.Metadata)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		guestVisit.Metadata = metadata
	}

	if err := handler.db.Save(&guestVisit).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update guest visit")
		return
	}

	if err := handler.db.Preload("Form").First(&guestVisit, "id = ?", guestVisit.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load updated guest visit")
		return
	}

	writeJSON(w, http.StatusOK, newGuestVisitResponse(guestVisit))
}

func (handler AdminGuestVisitHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestVisit, ok := handler.findGuestVisit(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	now := time.Now()
	guestVisit.Status = "checked_out"
	guestVisit.CheckOutAt = &now

	if err := handler.db.Save(&guestVisit).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check out guest visit")
		return
	}

	writeJSON(w, http.StatusOK, newGuestVisitResponse(guestVisit))
}

func (handler AdminGuestVisitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := authClaims(w, r)
	if !ok {
		return
	}

	guestVisit, ok := handler.findGuestVisit(w, claims.CompanyID, r.PathValue("id"))
	if !ok {
		return
	}

	if err := handler.db.Delete(&guestVisit).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete guest visit")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler AdminGuestVisitHandler) findGuestVisit(w http.ResponseWriter, companyID string, id string) (models.GuestVisit, bool) {
	var guestVisit models.GuestVisit
	if err := handler.db.
		Preload("Form").
		Where("id = ? AND company_id = ?", id, companyID).
		First(&guestVisit).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "guest visit not found")
			return models.GuestVisit{}, false
		}

		writeError(w, http.StatusInternalServerError, "failed to load guest visit")
		return models.GuestVisit{}, false
	}

	return guestVisit, true
}

func (handler AdminGuestVisitHandler) guestFormExists(companyID string, formID string) bool {
	var count int64
	return handler.db.Model(&models.GuestForm{}).
		Where("id = ? AND company_id = ?", formID, companyID).
		Count(&count).
		Error == nil && count > 0
}

func (request *guestVisitRequest) trim() {
	request.FormID = trimOptionalString(request.FormID)
	request.GuestName = trimOptionalString(request.GuestName)
	request.GuestEmail = trimOptionalString(request.GuestEmail)
	request.GuestPhone = trimOptionalString(request.GuestPhone)
	request.GuestCompany = trimOptionalString(request.GuestCompany)
	request.Purpose = trimOptionalString(request.Purpose)
	request.PersonToMeet = trimOptionalString(request.PersonToMeet)
	request.VisitDate = trimOptionalString(request.VisitDate)
	request.CheckOutAt = trimOptionalString(request.CheckOutAt)
	request.Status = normalizeGuestVisitStatus(request.Status)
	request.PhotoURL = trimOptionalString(request.PhotoURL)
	request.SignatureURL = trimOptionalString(request.SignatureURL)
}

func (request guestVisitRequest) validateCreate() error {
	if request.FormID == nil {
		return errors.New("form_id is required")
	}
	if request.GuestName == nil {
		return errors.New("guest_name is required")
	}
	if request.GuestPhone == nil {
		return errors.New("guest_phone is required")
	}
	if request.Purpose == nil {
		return errors.New("purpose is required")
	}

	return request.validateShared()
}

func (request guestVisitRequest) validateUpdate() error {
	return request.validateShared()
}

func (request guestVisitRequest) validateShared() error {
	if request.Status != nil && !validGuestVisitStatuses[*request.Status] {
		return errors.New("status is invalid")
	}

	return nil
}

func normalizeGuestVisitStatus(value *string) *string {
	if value == nil {
		return nil
	}

	status := strings.ToLower(strings.TrimSpace(*value))
	if status == "" {
		return nil
	}

	return &status
}

func parseOptionalDate(value *string) (time.Time, error) {
	if value == nil {
		return time.Time{}, nil
	}

	return parseDate(*value)
}

func parseDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func parseOptionalDateTime(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func newGuestVisitResponse(guestVisit models.GuestVisit) guestVisitResponse {
	return guestVisitResponse{
		ID:           guestVisit.ID,
		CompanyID:    guestVisit.CompanyID,
		FormID:       guestVisit.FormID,
		Form:         newGuestVisitForm(guestVisit.Form),
		GuestName:    guestVisit.GuestName,
		GuestEmail:   guestVisit.GuestEmail,
		GuestPhone:   guestVisit.GuestPhone,
		GuestCompany: guestVisit.GuestCompany,
		Purpose:      guestVisit.Purpose,
		PersonToMeet: guestVisit.PersonToMeet,
		VisitDate:    guestVisit.VisitDate.Format("2006-01-02"),
		CheckInAt:    guestVisit.CheckInAt,
		CheckOutAt:   guestVisit.CheckOutAt,
		Status:       guestVisit.Status,
		PhotoURL:     guestVisit.PhotoURL,
		SignatureURL: guestVisit.SignatureURL,
		Metadata:     guestVisit.Metadata,
		CreatedAt:    guestVisit.CreatedAt,
		UpdatedAt:    guestVisit.UpdatedAt,
	}
}

func newGuestVisitForm(guestForm models.GuestForm) guestVisitForm {
	return guestVisitForm{
		ID:         guestForm.ID,
		Name:       guestForm.Name,
		PublicSlug: guestForm.PublicSlug,
		Title:      guestForm.Title,
	}
}
