package user

import "time"

type User struct {
	// Id        string    `json:"id"` // Unused currently
	Username string `json:"username"`
	Password string `json:"password"`
	// Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
