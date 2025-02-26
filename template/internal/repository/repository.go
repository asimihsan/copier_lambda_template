package repository

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserCreate represents the data needed to create a user
type UserCreate struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserUpdate represents the data needed to update a user
type UserUpdate struct {
	Email *string `json:"email,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user UserCreate) (*User, error)
	UpdateUser(ctx context.Context, id string, user UserUpdate) (*User, error)
	DeleteUser(ctx context.Context, id string) error
}
