package repository

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `dynamodbav:"id" json:"id"`
	Email     string    `dynamodbav:"email" json:"email"`
	Name      string    `dynamodbav:"name" json:"name"`
	CreatedAt time.Time `dynamodbav:"created_at" json:"createdAt"`
	UpdatedAt time.Time `dynamodbav:"updated_at" json:"updatedAt"`
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
