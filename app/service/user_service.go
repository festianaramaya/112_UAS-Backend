package service

import (
	"database/sql"

	"uas/app/model"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create godoc
// @Summary Create new user
// @Description Create a new user (Admin only)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.CreateUserRequest true "Create User Request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := s.repo.Create(&user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User created successfully"})
}

// GetAll godoc
// @Summary Get all users
// @Description Get list of all users
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.User
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [get]
func (s *UserService) GetAll(c *fiber.Ctx) error {
	users, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	return c.JSON(users)
}

// GetDetail godoc
// @Summary Get user detail
// @Description Get detail of user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} model.User
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (s *UserService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.repo.GetUserByID(id)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user"})
	}

	return c.JSON(user)
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign role to a user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body object{role_id=string} true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id}/role [put]
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := s.repo.UpdateRole(userID, req.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user role"})
	}

	return c.JSON(fiber.Map{"message": "User role assigned successfully"})
}

// Delete godoc
// @Summary Delete user (soft delete)
// @Description Soft delete user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	userID := c.Params("id")

	if err := s.repo.SoftDelete(userID); err != nil {
		if err.Error() == "user not found or already deleted" {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

func (s *UserService) Update(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "User update not implemented yet",
	})
}
