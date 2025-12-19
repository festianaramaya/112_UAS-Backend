package routes

import (
	"uas/app/service"
	"uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authService *service.AuthService,
	userService *service.UserService,
	studentService *service.StudentService,
	lecturerService *service.LecturerService,
	achievementService *service.AchievementService,
	jwtSecret string,
) {

	app.Static("/uploads", "./uploads")

	authMiddleware := middleware.AuthRequired(jwtSecret)
	checkPerm := middleware.CheckPermission

	manageUser := checkPerm("user:manage")
	verifyPerm := checkPerm("achievement:verify")

	v1 := app.Group("/api/v1")

	// AUTH
	v1.Post("/auth/login", authService.Login)

	api := v1.Group("/", authMiddleware)
	api.Post("/auth/refresh", authService.Refresh)
	api.Post("/auth/logout", authService.Logout)
	api.Get("/auth/profile", authService.GetProfile)

	// USERS
	api.Get("/users", manageUser, userService.GetAll)
	api.Post("/users", manageUser, userService.Create)
	api.Get("/users/:id", manageUser, userService.GetDetail)
	api.Put("/users/:id", manageUser, userService.Update)
	api.Delete("/users/:id", manageUser, userService.Delete)
	api.Put("/users/:id/role", manageUser, userService.AssignRole)

	// STUDENTS
	api.Post("/students", manageUser, studentService.Create)
	api.Get("/students", studentService.GetAll)
	api.Get("/students/:id", studentService.GetDetail)
	api.Get("/students/:id/achievements", studentService.GetAchievements)
	api.Put("/students/:id/advisor", manageUser, studentService.SetAdvisor)

	// LECTURERS
	api.Get("/lecturers", lecturerService.GetAll)
	api.Get("/lecturers/:id/advisees", lecturerService.GetAdvisees)

	// ACHIEVEMENTS
	api.Get("/achievements", achievementService.GetAll)
	api.Get("/achievements/:id", achievementService.GetDetail)
	api.Post("/achievements", checkPerm("achievement:create"), achievementService.Create)
	api.Put("/achievements/:id", checkPerm("achievement:update"), achievementService.Update)
	api.Delete("/achievements/:id", checkPerm("achievement:delete"), achievementService.Delete)

	// WORKFLOW
	api.Post("/achievements/:id/submit", achievementService.Submit)
	api.Post("/achievements/:id/verify", verifyPerm, achievementService.Verify)
	api.Post("/achievements/:id/reject", verifyPerm, achievementService.Reject)

	// FILE & HISTORY
	api.Post("/achievements/:id/attachments", checkPerm("achievement:update"), achievementService.UploadAttachment)
	api.Get("/achievements/:id/history", achievementService.GetHistory)

	// REPORT
	api.Get("/reports/statistics", achievementService.GetStatistics)
	api.Get("/reports/student/:id", achievementService.GetStudentReport)
	
}
