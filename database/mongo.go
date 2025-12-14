package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo membuat koneksi ke MongoDB dan mengembalikan objek Database.
// Tipe kembalian *mongo.Database ini mengatasi error "mongoClient.Database undefined"
func ConnectMongo() (*mongo.Database, error) { 
	// 1. Ambil URI dan Nama Database dari environment
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DATABASE")

	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI or MONGO_DATABASE not set in environment variables.")
	}

	// 2. Setup Context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Buat Client Koneksi
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// 4. Ping Server untuk Verifikasi Koneksi
	err = client.Ping(ctx, nil)
	if err != nil {
		// Jika Ping gagal, tutup koneksi client
		if closeErr := client.Disconnect(context.Background()); closeErr != nil {
			log.Printf("Error closing MongoDB client after failed ping: %v", closeErr)
		}
		return nil, err
	}

	// 5. Pilih Database dan Kembalikan Objek Database
	// Ini adalah bagian yang paling penting untuk menyelesaikan error di main.go
	mongoDB := client.Database(dbName)

	log.Printf("Successfully connected to MongoDB Database: %s", dbName)

	return mongoDB, nil
}