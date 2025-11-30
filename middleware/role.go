package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// CheckPermission memverifikasi apakah user memiliki permission yang dibutuhkan.
func CheckPermission(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Dapatkan permissions dari Locals yang disetel oleh AuthRequired
		permsInterface := c.Locals("permissions")

		// Pastikan permission ditemukan
		if permsInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"error": "Forbidden: User permissions not loaded",
			})
		}

		userPermissions, ok := permsInterface.([]string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"error": "Internal Error: Cannot process permissions",
			})
		}

		// Cek apakah user memiliki permission yang diperlukan
		hasPermission := false
		for _, perm := range userPermissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"error": "Forbidden: Insufficient permissions for action: " + requiredPermission,
			})
		}

		return c.Next()
	}
}