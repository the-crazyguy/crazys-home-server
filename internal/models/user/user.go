package user

import (
	"errors"
	"time"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	// Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func ValidateUser(user *User) error {
	if user == nil {
		return errors.New("No user provided")
	}
	// TODO: Validate id as UUID
	if user.ID == "" {
		return errors.New("Empty ID")
	}

	if user.Username == "" {
		return errors.New("Empty username")
	}

	if user.Password == "" {
		return errors.New("Invalid password")
	}

	return nil
}
