package auth

import "golang.org/x/crypto/bcrypt"

const PasswordCost = 12

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), PasswordCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
