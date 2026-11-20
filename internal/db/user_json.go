package db

import (
	"encoding/json"
	"time"
)

// MarshalJSON keeps password hashes out of every API response using User.
// The generated database model still carries the hash for authentication.
func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID          int32     `json:"id"`
		Name        string    `json:"name"`
		Phone       string    `json:"phone"`
		Email       string    `json:"email"`
		BonusPoints int32     `json:"bonus_points"`
		CreatedAt   time.Time `json:"created_at"`
		IsActive    bool      `json:"is_active"`
		Role        string    `json:"role"`
	}{u.ID, u.Name, u.Phone, u.Email, u.BonusPoints, u.CreatedAt, u.IsActive, u.Role})
}
