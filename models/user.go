package models
import "time"

type User struct {
	ID string `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"-"` // The "-" ensures the hash is NEVER sent in JSON responses
	Name string `json:"full_name"`
	Role string `json:"role"` // "CUSTOMER" or "SELLER"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	Role string `json:"role" validate:"required,oneof=CUSTOMER SELLER"`
}