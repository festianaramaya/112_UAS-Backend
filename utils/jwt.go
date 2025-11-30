package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims -> FIX: Menambahkan Role (string) dan Perms ([]string) untuk RBAC Middleware.
// RoleID dipertahankan untuk lookup detail jika diperlukan.
type JWTClaims struct {
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
	Role     string `json:"role"` // Nama role (e.g., "Admin", "Mahasiswa")
	Perms    []string `json:"perms"`  // Daftar permissions (e.g., "achievement:create")
	jwt.RegisteredClaims
}

// GenerateToken -> FIX: Menerima nama role dan permissions untuk disimpan dalam token.
// Digunakan saat login.
func GenerateToken(userID, roleID, roleName string, permissions []string, secret string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RoleID: roleID,
		Role: roleName,  // Diisi dari database saat login
		Perms: permissions, // Diisi dari database saat login
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken -> Fungsi ini tidak berubah karena ia mem-parse structure JWTClaims
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	// Note: Menggunakan &JWTClaims{} untuk mendapatkan instance kosong sebagai target parsing
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Pastikan signing method yang digunakan sesuai
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}