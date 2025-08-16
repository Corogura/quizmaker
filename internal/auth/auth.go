package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytePW, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return "", err
	}
	return string(bytePW), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
