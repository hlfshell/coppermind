package users

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserAuth struct {
	ID                    string    `json:"id,omitempty" db:"id"`
	Password              string    `json:"password,omitempty" db:"password"`
	ResetToken            string    `json:"reset_token,omitempty" db:"reset_token"`
	ResetTokenAttempts    int       `json:"reset_token_attempts,omitempty" db:"reset_token_attempts"`
	ResetTokenGeneratedAt time.Time `json:"reset_token_generated_at,omitempty" db:"reset_token_generated_at"`
}

// CheckPasswordValidity will return true if the given password
// meets the minimum requirements for a password. For now
// that's just length but could be more in the future
func (userAuth *UserAuth) CheckPasswordValidity(password string) bool {
	return len(password) >= 8
}

// SetPassword checks the password for validity and then
// saves the salt and hashed output of the password to the
// UserAuth struct
func (userAuth *UserAuth) SetPassword(password string) error {
	if !userAuth.CheckPasswordValidity(password) {
		return fmt.Errorf("invalid password")
	}
	hash, err := generatePasswordHash(password)
	if err != nil {
		return err
	}

	userAuth.Password = hash
	return nil
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hash), err
}

func (userAuth *UserAuth) CheckPassword(given string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userAuth.Password), []byte(given))
	return err == nil
}

func (userAuth *UserAuth) Equal(other *UserAuth) bool {
	timeDifference := userAuth.ResetTokenGeneratedAt.Sub(other.ResetTokenGeneratedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}

	return userAuth.ID == other.ID &&
		userAuth.Password == other.Password &&
		userAuth.ResetToken == other.ResetToken &&
		userAuth.ResetTokenAttempts == other.ResetTokenAttempts &&
		timeDifference < time.Second
}

func (userAuth *UserAuth) GenerateResetToken() {
	userAuth.ResetToken = generateRandomString(8)
	userAuth.ResetTokenGeneratedAt = time.Now()
	userAuth.ResetTokenAttempts = 0
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
