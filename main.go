package main

import (
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    swagger "github.com/gofiber/swagger"
    _ "uas/docs"

    "github.com/joho/godotenv"
    "github.com/jmoiron/sqlx"

    "uas/database"
    "uas/routes"
    "uas/app/service"
    "uas/app/repository"
    "uas/utils"
)

// @title Achievement Management API
// @version 1.0
// @description API untuk sistem achievement mahasiswa
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer " diikuti token JWT kamu. Contoh: Bearer eyJhbGci...
func main() {
	_ = godotenv.Load()

	// DB
	pgStd, err := database.ConnectPostgres()
	if err != nil {
		log.Fatal(err)
	}
	defer pgStd.Close()

	pgDB := sqlx.NewDb(pgStd, "postgres")
	_ = utils.SetupDatabase(pgStd)

	mongoDB, _ := database.ConnectMongo()
	achievementCollection := mongoDB.Collection("achievements")

	// ENV
	jwtSecret := os.Getenv("JWT_SECRET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Repository
	userRepo := repository.NewUserRepository(pgDB)
	studentRepo := repository.NewStudentRepository(pgDB)
	lecturerRepo := repository.NewLecturerRepository(pgDB)
	pgAchievementRepo := repository.NewAchievementRepository(pgDB)
	mongoAchievementRepo := repository.NewMongoAchievementRepository(achievementCollection)

	// Service
	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	studentService := service.NewStudentService(studentRepo, pgAchievementRepo)
	lecturerService := service.NewLecturerService(pgAchievementRepo, lecturerRepo)
	achievementService := service.NewAchievementService(pgAchievementRepo, mongoAchievementRepo)

	// App
	app := fiber.New()

	// Swagger (INI YANG BENAR)
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	routes.SetupRoutes(
		app,
		authService,
		userService,
		studentService,
		lecturerService,
		achievementService,
		jwtSecret,
	)

	log.Fatal(app.Listen(":" + port))
}
