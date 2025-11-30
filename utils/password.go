package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    // menggunakan bcrypt dari golang.org/x/crypto/bcrypt
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(input string, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input))
	return err == nil
}
