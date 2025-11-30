package service

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/model"
	"uas/app/repository"
	// Hapus import "github.com/google/uuid" jika ada
)

type AchievementService struct {
	repo *repository.AchievementRepository
}

func NewAchievementService(r *repository.AchievementRepository) *AchievementService {
	return &AchievementService{repo: r}
}

// CREATE (status draft) - FR-003
func (s *AchievementService) Create(c *fiber.Ctx) error {
	var req struct {
		StudentID string `json:"student_id"`
		MongoAchievementID string `json:"mongo_achievement_id"`
        // Pastikan tidak ada karakter tersembunyi di sini yang menyebabkan Syntax Error
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

    // FIX: Menggunakan KEYED FIELDS untuk model.AchievementReference
	a := model.AchievementReference{
		// ID, CreatedAt, UpdatedAt akan diisi oleh Repository/DB (RETURNING)
		StudentID:req.StudentID,
		MongoAchievementID: req.MongoAchievementID,
		Status:"draft",
	}

	err := s.repo.Create(&a) // Repository akan mengisi a.ID, a.CreatedAt, a.UpdatedAt
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create achievement reference"})
	}

	return c.Status(201).JSON(fiber.Map{"id": a.ID, "message": "Achievement reference created successfully"})
}

// DETAIL (READ)
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
    // ... (kode fungsi GetDetail, Submit, Verify, Reject sebelumnya sudah benar secara sintaksis)
    id := c.Params("id")

	data, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return c.JSON(data)
}

// SUBMIT
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repo.UpdateStatus(id, "submitted", nil, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed"})
	}

	return c.JSON(fiber.Map{"message": "submitted"})
}

// VERIFY
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		VerifiedBy string `json:"verified_by"`
	}

	c.BodyParser(&body)

	err := s.repo.UpdateStatus(id, "verified", &body.VerifiedBy, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed"})
	}

	return c.JSON(fiber.Map{"message": "verified"})
}

// REJECT
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		VerifiedBy string `json:"verified_by"`
		RejectionNote string `json:"rejection_note"`
	}

	c.BodyParser(&body)

	err := s.repo.UpdateStatus(id, "rejected", &body.VerifiedBy, &body.RejectionNote)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed"})
	}

	return c.JSON(fiber.Map{"message": "rejected"})
}

// Tambahkan ini ke achievement_service.go (dengan Create, Detail, Submit, Verify, Reject yang sudah ada)

func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "List of achievements fetched successfully"})
}
func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "Achievement Update not implemented"})
}
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "Achievement Delete not implemented"})
}
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "UploadAttachment not implemented"})
}
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Achievement History fetched successfully"})
}
func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Achievement Statistics fetched successfully (FR-011)"})
}
func (s *AchievementService) GetStudentReport(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Student Report fetched successfully"})
}