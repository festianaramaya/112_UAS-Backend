package service

import (
	"database/sql"
    "uas/app/repository"
    "uas/utils" // FIX: Tambahkan import utils
    "uas/app/model" // FIX: Tambahkan import model
    
    "github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(c *fiber.Ctx) error {
    var req model.CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    // 1. HASH PASSWORD
    passwordHash, err := utils.HashPassword(req.Password)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    // 2. Map ke Model User
    user := model.User{
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: passwordHash, 
        FullName:     req.FullName,
        RoleID:       req.RoleID,
        IsActive:     true,
    }

    // 3. Kirim ke Repository (Membutuhkan s.repo.Create)
    if err := s.repo.Create(&user); err != nil { // ERROR s.repo.Create akan hilang setelah Langkah 2
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
    }

    return c.Status(201).JSON(fiber.Map{"message": "User created successfully"})
}

func (s *UserService) GetAll(c *fiber.Ctx) error {
    // ... (kode ini sudah benar)
	users, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	return c.JSON(users)
}

// GetDetail - ID user adalah UUID (string), bukan integer
func (s *UserService) GetDetail(c *fiber.Ctx) error {
	idParam := c.Params("id") // ID user adalah string (UUID)

	// FIX: Hapus konversi ke int (strconv.Atoi)
	// id, err := strconv.Atoi(idParam) <-- HAPUS BARIS INI
	// if err != nil { ... }

	user, err := s.repo.GetUserByID(idParam) // Langsung gunakan idParam (string)
	
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user"})
	}

	return c.JSON(user)
}

func (s *UserService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "User Update (Admin) not implemented"})
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
    // Memerlukan role_id dan id user dari parameter URL
    userID := c.Params("id") 

    var req struct {
        RoleID string `json:"role_id"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // 1. Validasi RoleID (Opsional: Pastikan RoleID ada di tabel roles)
    // Logika ini bisa ditambahkan di sini atau di repository.

    // 2. Update role di repository
    if err := s.repo.UpdateRole(userID, req.RoleID); err != nil {
        // Asumsi error spesifik untuk not found ditangkap di repository
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user role"})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User role assigned successfully"})
}

// Delete (DELETE /api/v1/users/:id) - Melakukan Soft Delete
func (s *UserService) Delete(c *fiber.Ctx) error {
    userID := c.Params("id") // Ambil ID dari parameter URL

    // 1. Panggil Repository untuk melakukan Soft Delete
    if err := s.repo.SoftDelete(userID); err != nil {
        
        // 2. Tangani error spesifik jika user tidak ditemukan/sudah dihapus
        if err.Error() == "user not found or already deleted" {
             return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found or already deleted"})
        }
        
        // 3. Tangani error database umum
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user"})
    }

    // 4. Kembalikan status sukses
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully (soft delete)"})
}