package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
    "go.mongodb.org/mongo-driver/mongo" // <--- DIBUTUHKAN
	"go.mongodb.org/mongo-driver/mongo/options" // <--- DIBUTUHKAN
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

// ConnectMongo menginisialisasi koneksi ke MongoDB (mengambil URI dari .env)
func ConnectMongo() (*mongo.Database, error) {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB")
	
	if mongoURI == "" || mongoDBName == "" {
		log.Fatal("MONGO_URI or MONGO_DB environment variables are not set.")
	}
	
	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Println("❌ Gagal koneksi MongoDB client:", err)
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Println("❌ Gagal ping MongoDB:", err)
		return nil, err
	}

	log.Println("✅ Berhasil konek ke database MongoDB.")
	return client.Database(mongoDBName), nil
}