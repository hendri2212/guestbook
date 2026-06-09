package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"guestbook/backend/models"
	"guestbook/backend/router"
	"guestbook/backend/seeders"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	loadEnv()

	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := autoMigrate(db); err != nil {
		log.Fatalf("failed to run auto migration: %v", err)
	}

	if err := seeders.SeedDefaultData(db); err != nil {
		log.Fatalf("failed to seed default data: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("guestbook backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router.New(db, mysqlPort())))
}

func loadEnv() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("failed to load .env: %v", err)
	}
}

func connectDatabase() (*gorm.DB, error) {
	config := mysqlConfigFromEnv()

	if config.DSN == "" {
		if err := ensureDatabase(config); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(mysql.Open(config.dsn()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Company{},
		&models.AdminUser{},
		&models.GuestForm{},
		&models.GuestVisit{},
	)
}

type mysqlConfig struct {
	DSN      string
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func mysqlConfigFromEnv() mysqlConfig {
	return mysqlConfig{
		DSN:      os.Getenv("MYSQL_DSN"),
		User:     envOrDefault("MYSQL_USER", "root"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Host:     envOrDefault("MYSQL_HOST", "127.0.0.1"),
		Port:     mysqlPort(),
		Database: envOrDefault("MYSQL_DATABASE", "guestbook"),
	}
}

func (config mysqlConfig) dsn() string {
	if config.DSN != "" {
		return config.DSN
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
}

func (config mysqlConfig) serverDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
	)
}

func mysqlPort() string {
	return envOrDefault("MYSQL_PORT", "8889")
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func ensureDatabase(config mysqlConfig) error {
	db, err := sql.Open("mysql", config.serverDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec(
		fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			escapeIdentifier(config.Database),
		),
	)

	return err
}

func escapeIdentifier(identifier string) string {
	return strings.ReplaceAll(identifier, "`", "``")
}
