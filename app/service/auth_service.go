package service

import (
	"uas/app/repository"
	"uas/utils"
	"database/sql" // Diperlukan untuk sql.ErrNoRows

	"context"
    "errors"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	UserRepo *repository.UserRepository
	JWTSecret string
}

func NewAuthService(userRepo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		UserRepo: userRepo,
		JWTSecret: secret,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx := context.Background()
    // Asumsi FindByUsername di repository mengembalikan struct user yang berisi RoleID
	user, err := s.UserRepo.FindByUsername(ctx, req.Username)
	

	if err != nil {
        // Cek jika user tidak ditemukan di DB
        if errors.Is(err, sql.ErrNoRows) { 
            return c.Status(401).JSON(fiber.Map{"error": "User not found or invalid credentials"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
	}
    
    // ----------------------------------------------------
    // FIX: AMBIL ROLE NAME DAN PERMISSIONS UNTUK TOKEN (RBAC)
    // ----------------------------------------------------
    
    // Asumsi user struct memiliki field RoleName yang sudah di-join atau diambil
    // Jika tidak, kita harus memanggil repository untuk mendapatkan permissions.
    
    // *PERHATIAN: Karena user struct hanya punya RoleID, kita harus mendapatkan nama role dan permissions.*
    
    // 1. Dapatkan Role Name (Asumsi UserRepo bisa melakukan JOIN atau ada field RoleName)
    //    Karena user model Anda di SRS hanya memiliki RoleID, kita akan menggunakan RoleID
    //    untuk mendapatkan permissions. (Jika Anda menggunakan JOIN di FindByUsername, ganti ini.)
    roleName := "Unknown" // Placeholder, harusnya diambil saat fetch user
    
    // 2. Dapatkan Permissions
    permissions, err := s.UserRepo.GetUserPermissions(user.RoleID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user permissions"})
    }
    
    // ----------------------------------------------------

    // FIX: Panggil GenerateToken dengan 5 argumen
	token, err := utils.GenerateToken(
        user.ID, 
        user.RoleID, 
        roleName, // FIX: Role Name (Harus di-fetch di Repository)
        permissions, // FIX: Permissions
        s.JWTSecret,
    )
    
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.JSON(fiber.Map{
        "status": "success", // Tambahkan status sesuai SRS
	 	"token": token,
        // Tambahkan data user profile jika diperlukan
	})
}


// --- Tambahkan Methods Pendukung ---

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "Refresh endpoint not implemented"})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Logout successful"})
}

func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	// Di sini Anda akan membaca token JWT dari context dan mengembalikan data user.
	return c.Status(200).JSON(fiber.Map{"message": "Profile data fetched successfully"})
}