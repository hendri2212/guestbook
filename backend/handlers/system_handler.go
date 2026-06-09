package handlers

import (
	"net/http"
	"time"

	"gorm.io/gorm"
)

type SystemHandler struct {
	db        *gorm.DB
	mysqlPort string
}

type healthResponse struct {
	App       string    `json:"app"`
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

type migrationResponse struct {
	Database string   `json:"database"`
	Port     string   `json:"port"`
	Models   []string `json:"models"`
}

func NewSystemHandler(db *gorm.DB, mysqlPort string) SystemHandler {
	return SystemHandler{db: db, mysqlPort: mysqlPort}
}

func (handler SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := pingDatabase(handler.db); err != nil {
		status = "database_error"
	}

	writeJSON(w, http.StatusOK, healthResponse{
		App:       "guestbook-backend",
		Status:    status,
		Database:  "mysql",
		Timestamp: time.Now().UTC(),
	})
}

func (handler SystemHandler) Migrations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, migrationResponse{
		Database: "mysql",
		Port:     handler.mysqlPort,
		Models: []string{
			"companies",
			"admin_users",
			"guest_forms",
			"guest_visits",
		},
	})
}

func pingDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
