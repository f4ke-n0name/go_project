package models

import "time"

type Role string

const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "manager"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
