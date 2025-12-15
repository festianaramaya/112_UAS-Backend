package routes

import (
	"uas/app/service"
	"uas/middleware"

	"github.com/gofiber/fiber/v2"
)

// Pastikan semua service yang dibutuhkan dideklarasikan di service layer.
func SetupRoutes(
	authService *service.AuthService,
	userService *service.UserService,
	studentService *service.StudentService,
	lecturerService *service.LecturerService,
	achievementService *service.AchievementService,
	jwtSecret string, // Digunakan untuk AuthMiddleware
) *fiber.App {

	app := fiber.New()

	app.Static("/uploads", "./uploads")
	
	// FIX: Inisialisasi Auth Middleware (Asumsi middleware.AuthRequired mengembalikan fiber.Handler)
	authMiddleware := middleware.AuthRequired(jwtSecret) 
	checkPerm := middleware.CheckPermission
	
	// Permissions
	manageUser := checkPerm("user:manage")
	verifyPerm := checkPerm("achievement:verify")

	v1 := app.Group("/api/v1") 

	// AUTHENTICATION (PUBLIC & PROTECTED)
	v1.Post("/auth/login", authService.Login) 
	
	// FIX: Group Protected API
	api := v1.Group("/", authMiddleware)
	
	api.Post("/auth/refresh", authService.Refresh) // Method Refresh harus ada di AuthService
	api.Post("/auth/logout", authService.Logout)   // Method Logout harus ada di AuthService
	api.Get("/auth/profile", authService.GetProfile) // Method GetProfile harus ada di AuthService

	// USERS (Admin Management)
	api.Get("/users", manageUser, userService.GetAll)          
	api.Post("/users", manageUser, userService.Create)         // FIX: Method Create harus ada
	api.Get("/users/:id", manageUser, userService.GetDetail)    
	api.Put("/users/:id", manageUser, userService.Update)       // FIX: Method Update harus ada
	api.Delete("/users/:id", manageUser, userService.Delete)    // FIX: Method Delete harus ada
	api.Put("/users/:id/role", manageUser, userService.AssignRole) // FIX: Method AssignRole harus ada

	// STUDENTS & LECTURERS
	api.Get("/students", studentService.GetAll)
	api.Get("/students/:id", studentService.GetDetail)
	api.Get("/students/:id/achievements", studentService.GetAchievements) // FIX: Method GetAchievements harus ada
	api.Put("/students/:id/advisor", manageUser, studentService.SetAdvisor) // FIX: Method SetAdvisor harus ada

	api.Get("/lecturers", lecturerService.GetAll)
	api.Get("/lecturers/:id/advisees", lecturerService.GetAdvisees) // FIX: Method GetAdvisees harus ada
	
	// ACHIEVEMENTS
	api.Get("/achievements", achievementService.GetAll) // FIX: Method GetAll harus ada
	api.Get("/achievements/:id", achievementService.GetDetail)
	api.Post("/achievements", checkPerm("achievement:create"), achievementService.Create) 
	api.Put("/achievements/:id", checkPerm("achievement:update"), achievementService.Update) // FIX: Method Update harus ada
	api.Delete("/achievements/:id", checkPerm("achievement:delete"), achievementService.Delete) // FIX: Method Delete harus ada

	// Workflow
	api.Post("/achievements/:id/submit", achievementService.Submit) 
	api.Post("/achievements/:id/verify", verifyPerm, achievementService.Verify) 
	api.Post("/achievements/:id/reject", verifyPerm, achievementService.Reject) 
	
	// File & History
	api.Post("/achievements/:id/attachments", checkPerm("achievement:update"), achievementService.UploadAttachment) // FIX: Method UploadAttachment harus ada
	api.Get("/achievements/:id/history", achievementService.GetHistory) // FIX: Method GetHistory harus ada
	
	// REPORTS & ANALYTICS
	api.Get("/reports/statistics", achievementService.GetStatistics) // FIX: Method GetStatistics harus ada
	api.Get("/reports/student/:id", achievementService.GetStudentReport) // FIX: Method GetStudentReport harus ada

	return app
}