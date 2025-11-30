package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"uas/database"
	"uas/routes"
	"uas/app/service"
	"uas/app/repository"
	"uas/utils" // Import utilitas untuk seeder

)

func main() {
	// 1. LOAD .ENV FILE
	godotenv.Load()

	// 2. CONNECT DATABASE (PostgreSQL)
	pgDB, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("Failed to connect PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	// 3. DDL & SEEDING 
	if err := utils.SetupDatabase(pgDB); err != nil {
		log.Fatalf("Failed to setup database (DDL/Seeding): %v", err)
	}

	// 4. CONNECT DATABASE (MongoDB)
	_, err = database.ConnectMongo() 
	if err != nil {
		log.Fatalf("Failed to connect MongoDB: %v", err)
	}
	
	// 5. AMBIL CONFIG DARI ENV
	jwtSecret := os.Getenv("JWT_SECRET")
	appPort := os.Getenv("PORT")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}
	if appPort == "" {
		appPort = "3000"
	}


	// REPOSITORY (DI)
	userRepo := repository.NewUserRepository(pgDB)
	// Placeholder untuk Repository lain
	studentRepo := &repository.StudentRepository{}
	achievementRepo := &repository.AchievementRepository{}
	// FIX: Hapus '&' yang tidak diperlukan
	lecturerRepo := repository.NewLecturerRepository(pgDB) 

	// SERVICE (DI)
	authService := service.NewAuthService(userRepo, jwtSecret) 
	userService := service.NewUserService(userRepo)
	studentService := service.NewStudentService(studentRepo)
	achievementService := service.NewAchievementService(achievementRepo)
	lecturerService := service.NewLecturerService(lecturerRepo)

	// ROUTES
	// FIX: Sesuaikan urutan pemanggilan dengan signature di routes.go
	app := routes.SetupRoutes(
		authService,
		userService,
		studentService,
		lecturerService,
		achievementService,
		jwtSecret, // JWT Secret sebagai parameter terakhir
	)

	// LISTEN
	log.Printf("Server running on port %s", appPort)
	log.Fatal(app.Listen(":" + appPort))
}