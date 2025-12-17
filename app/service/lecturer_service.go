package service

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"uas/app/model"
	"uas/app/repository"
)

type LecturerService struct {
	MongoAchieveRepo *repository.MongoAchievementRepository
	LecturerRepo     *repository.LecturerRepository
}

// helper parse UUID
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

// GET ALL lecturers
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

// GET advisee achievements (MONGO)
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID, err := parseUUIDParam(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := context.Background()

	results, err := s.MongoAchieveRepo.GetAdviseeAchievements(ctx, lecturerID)
	if err != nil {
		log.Println("Mongo GetAdvisees error:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(results) == 0 {
		return c.JSON(fiber.Map{
			"message": "No advisee achievements found",
			"data":    []model.AchievementFull{},
			"total":   0,
		})
	}

	return c.JSON(fiber.Map{
		"message": "List of advisee achievements",
		"data":    results,
		"total":   len(results),
	})
}

// constructor (MONGO VERSION)
func NewLecturerService(
	mongoAchieveRepo *repository.MongoAchievementRepository,
	lecturerRepo *repository.LecturerRepository,
) *LecturerService {
	return &LecturerService{
		MongoAchieveRepo: mongoAchieveRepo,
		LecturerRepo:     lecturerRepo,
	}
}
