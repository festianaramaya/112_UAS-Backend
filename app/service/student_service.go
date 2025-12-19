package service

import (
	"context"
	"database/sql"
	"log"

	"uas/app/model"
	"uas/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentService struct {
	StudentRepo     *repository.StudentRepository
	AchievementRepo *repository.AchievementRepository
}

func NewStudentService(
	studentRepo *repository.StudentRepository,
	achievementRepo *repository.AchievementRepository,
) *StudentService {
	return &StudentService{
		StudentRepo:     studentRepo,
		AchievementRepo: achievementRepo,
	}
}

// Create godoc
// @Summary Create student
// @Description Create new student
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.Student true "Student Data"
// @Success 201 {object} model.Student
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/students [post]
func (s *StudentService) Create(c *fiber.Ctx) error {
	var req model.Student

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx := context.Background()

	if err := s.StudentRepo.Create(ctx, &req); err != nil {
		log.Printf("ERROR create student: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create student"})
	}

	return c.Status(201).JSON(req)
}

// GetAll godoc
// @Summary Get all students
// @Description Get list of students
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.Student
// @Failure 500 {object} map[string]string
// @Router /api/v1/students [get]
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	ctx := context.Background()

	result, err := s.StudentRepo.GetAll(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed fetching students"})
	}

	return c.JSON(result)
}

// GetDetail godoc
// @Summary Get student detail
// @Description Get student detail by ID
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} model.Student
// @Failure 404 {object} map[string]string
// @Router /api/v1/students/{id} [get]
func (s *StudentService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	ctx := context.Background()
	result, err := s.StudentRepo.GetStudentByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	return c.JSON(result)
}

// GetAchievements godoc
// @Summary Get student achievements
// @Description Get all achievements of a student
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/students/{id}/achievements [get]
func (s *StudentService) GetAchievements(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	data, err := s.AchievementRepo.GetByStudentID(
		c.Context(),
		studentID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed fetching achievements"})
	}

	return c.JSON(fiber.Map{
		"student_id": studentID,
		"total":      len(data),
		"data":       data,
	})
}

// SetAdvisor godoc
// @Summary Set student advisor
// @Description Assign advisor (dosen wali) to student
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param request body object{advisor_id=string} true "Advisor ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/students/{id}/advisor [put]
func (s *StudentService) SetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")

	var req struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var advisorID sql.NullString

	if req.AdvisorID != "" {
		if _, err := uuid.Parse(req.AdvisorID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid advisor UUID"})
		}
		advisorID.String = req.AdvisorID
		advisorID.Valid = true
	}

	ctx := context.Background()

	if err := s.StudentRepo.UpdateAdvisorID(ctx, studentID, advisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update advisor"})
	}

	return c.JSON(fiber.Map{"message": "Student advisor updated successfully"})
}
