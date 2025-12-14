package database

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// ConnectPostgres menginisialisasi koneksi ke PostgreSQL
func ConnectPostgres() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set.")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("❌ Gagal membuka koneksi DB:", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Println("❌ Gagal ping DB:", err)
		return nil, err
	}

	// Konfigurasi Pooling dari .env
	if maxConnStr := os.Getenv("DB_MAX_CONNECTIONS"); maxConnStr != "" {
		if maxConn, err := strconv.Atoi(maxConnStr); err == nil {
			db.SetMaxOpenConns(maxConn)
		}
	}
	if maxIdleStr := os.Getenv("DB_MAX_IDLE_CONNECTIONS"); maxIdleStr != "" {
		if maxIdle, err := strconv.Atoi(maxIdleStr); err == nil {
			db.SetMaxIdleConns(maxIdle)
		}
	}
	if maxLifetimeStr := os.Getenv("DB_MAX_LIFETIME_CONNECTIONS"); maxLifetimeStr != "" {
		if duration, err := time.ParseDuration(maxLifetimeStr); err == nil {
			db.SetConnMaxLifetime(duration)
		}
	}

	log.Println("✅ Berhasil konek ke database PostgreSQL.")
	return db, nil
}