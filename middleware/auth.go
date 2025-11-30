package middleware

import (
	"strings"
	"uas/utils" // Asumsi utils.ParseToken dan claims structs ada di sini

	"github.com/gofiber/fiber/v2"
)

// AuthRequired mengembalikan fiber.Handler yang memverifikasi JWT.
// Fungsi ini menerima JWT Secret saat inisialisasi (closure).
func AuthRequired(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"error": "Unauthorized: Missing or invalid token format",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse token menggunakan secret yang di-closure
		claims, err := utils.ParseToken(tokenString, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"error": "Unauthorized: Invalid or expired token",
			})
		}

		// Simpan data user dari Claims ke Locals (diperlukan untuk CheckPermission)
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role) // Asumsi claims.Role ada untuk RBAC
		c.Locals("permissions", claims.Perms) // Asumsi claims.Perms ada

		return c.Next()
	}
}