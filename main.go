package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jmoiron/sqlx"

	"uas/database"
	"uas/routes"
	"uas/app/service"
	"uas/app/repository"
	"uas/utils"
)

func main() {
	// Load environment variables dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Koneksi ke database PostgreSQL
	pgDBStandard, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("Failed to connect PostgreSQL: %v", err)
	}
	defer pgDBStandard.Close()

	// Konversi *sql.DB ke *sqlx.DB agar mendukung fitur sqlx
	pgDB := sqlx.NewDb(pgDBStandard, "postgres")

	// Menjalankan DDL dan seeding database
	if err := utils.SetupDatabase(pgDBStandard); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Koneksi ke MongoDB (mengembalikan *mongo.Database)
	mongoDB, err := database.ConnectMongo()
	if err != nil {
		log.Fatalf("Failed to connect MongoDB: %v", err)
	}

	// Mengambil collection achievements dari MongoDB
	achievementCollection := mongoDB.Collection("achievements")

	// Mengambil konfigurasi dari environment
	jwtSecret := os.Getenv("JWT_SECRET")
	appPort := os.Getenv("PORT")

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}

	if appPort == "" {
		appPort = "3000"
	}

	// Inisialisasi repository PostgreSQL
	userRepo := repository.NewUserRepository(pgDB)
	studentRepo := repository.NewStudentRepository(pgDB)
	lecturerRepo := repository.NewLecturerRepository(pgDB)

	// Inisialisasi repository achievement (Postgres & MongoDB)
	pgAchievementRepo := repository.NewAchievementRepository(pgDB)
	mongoAchievementRepo := repository.NewMongoAchievementRepository(achievementCollection)

	// Inisialisasi service
	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	studentService := service.NewStudentService(studentRepo)
	lecturerService := service.NewLecturerService(lecturerRepo)

	// Service achievement membutuhkan dua repository
	achievementService := service.NewAchievementService(
		pgAchievementRepo,
		mongoAchievementRepo,
	)

	// Setup routes dan middleware
	app := routes.SetupRoutes(
		authService,
		userService,
		studentService,
		lecturerService,
		achievementService,
		jwtSecret,
	)

	// Menjalankan server
	log.Printf("Server running on port %s", appPort)
	log.Fatal(app.Listen(":" + appPort))
}
