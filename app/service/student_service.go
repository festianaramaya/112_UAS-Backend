package service

import (
	"context"
	"uas/app/model"
	"uas/app/repository"

	"log"
	"database/sql"
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
    // Catatan: Model req (model.Student) tidak boleh lagi memiliki UpdatedAt field.

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// DB akan mengisinya via RETURNING
	ctx := context.Background()
	// Repository Create akan mengisi ID dan CreatedAt (dan HANYA ITU)
	if err := s.StudentRepo.Create(ctx, &req); err != nil { 
		log.Printf("ERROR: Failed to create student in repo: %v", err)
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

// GetAll - Mengambil daftar semua mahasiswa
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	ctx := context.Background()

	result, err := s.StudentRepo.GetAll(ctx) 
	if err != nil {
		log.Printf("ERROR: StudentRepo GetAll failed: %v", err) 
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
    studentID := c.Params("id") // ID profil students

    var req struct {
        AdvisorID string `json:"advisor_id"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // 1. Validasi ID yang di-request
    // Jika req.AdvisorID kosong, kita set ke NULL di database.
    var advisorIDToSet sql.NullString
    if req.AdvisorID != "" {
        // Cek apakah ID Dosen Wali adalah UUID yang valid
        // (Contoh validasi sederhana: memastikan panjang)
        if len(req.AdvisorID) != 36 {
             return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Advisor UUID format"})
        }
        advisorIDToSet.String = req.AdvisorID
        advisorIDToSet.Valid = true
    } else {
        advisorIDToSet.Valid = false // Set ke NULL
    }
    
    ctx := context.Background()

    // 2. Panggil Repository untuk melakukan UPDATE
    err := s.StudentRepo.UpdateAdvisorID(ctx, studentID, advisorIDToSet) 
    if err != nil {
        // Jika terjadi error, kemungkinan: Student ID tidak ditemukan atau Foreign Key (AdvisorID) tidak valid
        log.Printf("ERROR: Failed to update advisor for student %s: %v", studentID, err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update advisor ID. Check if Student ID or Advisor ID is valid."})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Student advisor updated successfully",
        "student_id": studentID,
        "new_advisor_id": advisorIDToSet.String,
    })
}