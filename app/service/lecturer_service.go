package service

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"uas/app/repository"
)

type LecturerService struct {
	AchievementRepo *repository.AchievementRepository
	LecturerRepo    *repository.LecturerRepository
}

// =========================
// helper parse UUID
// =========================
func parseUUIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	idStr := c.Params("id")
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("ID is required")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format")
	}

	return id, nil
}

// =========================
// GET /lecturers
// =========================
// GetAll godoc
// @Summary      Get all lecturers
// @Description  Mengambil semua daftar dosen yang ada di sistem
// @Tags         Lecturer
// @Accept       json
// @Produce      json
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /api/v1/lecturers [get]
// @Security     BearerAuth
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	lecturers, err := s.LecturerRepo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch lecturers",
		})
	}

	return c.JSON(fiber.Map{
		"message": "List of lecturers",
		"data":    lecturers,
		"total":   len(lecturers),
	})
}

// GetAdvisees godoc
// @Summary      Get list of advisee achievements
// @Description  Melihat daftar prestasi mahasiswa bimbingan (FR-006)
// @Tags         Lecturer
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/lecturers/{id}/advisees [get]
// @Security     BearerAuth
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID, err := parseUUIDParam(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := context.Background()

	results, err := s.AchievementRepo.GetByLecturerID(ctx, lecturerID)
	if err != nil {
		log.Println("GetAdvisees error:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch advisee achievements",
		})
	}

	if len(results) == 0 {
		return c.JSON(fiber.Map{
			"message": "No advisee achievements found",
			"data":    []any{},
			"total":   0,
		})
	}

	return c.JSON(fiber.Map{
		"message": "List of advisee achievements",
		"data":    results,
		"total":   len(results),
	})
}

// =========================
// CONSTRUCTOR
// =========================
func NewLecturerService(
	achievementRepo *repository.AchievementRepository,
	lecturerRepo *repository.LecturerRepository,
) *LecturerService {
	return &LecturerService{
		AchievementRepo: achievementRepo,
		LecturerRepo:    lecturerRepo,
	}
}
