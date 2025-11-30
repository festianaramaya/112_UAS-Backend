package service

import (
	"context"
	"uas/app/model"
	"uas/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	StudentRepo *repository.StudentRepository
}

func NewStudentService(repo *repository.StudentRepository) *StudentService {
	return &StudentService{
		StudentRepo: repo,
	}
}

// ===============================================
// Implementasi CRUD
// ===============================================

// Create (POST /students)
func (s *StudentService) Create(c *fiber.Ctx) error { 
	var req model.Student

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// DB akan mengisinya via RETURNING
	ctx := context.Background()
	if err := s.StudentRepo.Create(ctx, &req); err != nil { 
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create student"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Student created successfully",
		"data":req, 
	})
}

// GetDetail (GET /students/:id)
func (s *StudentService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	ctx := context.Background()
	result, err := s.StudentRepo.GetStudentByID(ctx, id) 
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	return c.JSON(result)
}

// GetAll (GET /students)
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	ctx := context.Background()

	result, err := s.StudentRepo.GetAll(ctx) 
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed fetching students"})
	}

	return c.JSON(result)
}


// ===============================================
// Implementasi METHOD LAIN (yang hilang di routes.go)
// ===============================================

// GetAchievements (GET /students/:id/achievements)
func (s *StudentService) GetAchievements(c *fiber.Ctx) error {
    studentID := c.Params("id")
    // Logika: 1. Cek keberadaan studentID, 2. Ambil list achievement references dari PostgreSQL, 
    // 3. Ambil detail achievement dari MongoDB.
	return c.Status(501).JSON(fiber.Map{"message": "GetAchievements not yet implemented for ID: " + studentID})
}

// SetAdvisor (PUT /students/:id/advisor) - Hanya Admin/manageUser
func (s *StudentService) SetAdvisor(c *fiber.Ctx) error {
    // Logika: Update field advisor_id di tabel students
	return c.Status(501).JSON(fiber.Map{"message": "SetAdvisor not yet implemented"})
}