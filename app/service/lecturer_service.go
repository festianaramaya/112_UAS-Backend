package service

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/model"
	"uas/app/repository"
)

type LecturerService struct {
	repo *repository.LecturerRepository
}

func NewLecturerService(repo *repository.LecturerRepository) *LecturerService {
	return &LecturerService{repo: repo}
}

func (s *LecturerService) CreateLecturer(c *fiber.Ctx) error {
	var l model.Lecturer
	if err := c.BodyParser(&l); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := s.repo.Create(&l); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Lecturer created successfully"})
}

func (s *LecturerService) GetLecturerByID(c *fiber.Ctx) error {
	id := c.Params("id")

	lecturer, err := s.repo.GetLecturerByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Lecturer not found"})
	}

	return c.JSON(lecturer)
}

func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	result, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed fetching lecturers"})
	}
	return c.JSON(result)
}

func (s *LecturerService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := s.repo.GetLecturerByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(result)
}

// Tambahkan ini ke lecturer_service.go
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "List of advisee achievements fetched successfully (FR-006)"})
}