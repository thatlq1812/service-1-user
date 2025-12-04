package repository

import (
	"context"
	pb "service-1-user/proto"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetByID user by ID
	GetByID(ctx context.Context, id int32) (*pb.User, error)

	// GetByEmailWithPassword user by email with password hash (for authentication)
	GetByEmailWithPassword(ctx context.Context, email string) (*UserWithPassword, error)

	// Create new user with password
	CreateWithPassword(ctx context.Context, name, email, passwordHash string) (*pb.User, error)

	// Create new user (legacy method without password)
	Create(ctx context.Context, name, email string) (*pb.User, error)

	// Update user information (full update)
	Update(ctx context.Context, id int32, name, email string) (*pb.User, error)

	// PartialUpdate user information (only provided fields)
	PartialUpdate(ctx context.Context, id int32, name *string, email *string, password *string) (*pb.User, error)

	// Delete user by ID
	Delete(ctx context.Context, id int32) error

	// List all user with pagination
	List(ctx context.Context, limit, offset int32) ([]*pb.User, int32, error)
}

// UserWithPassword extends User with password_hash field for internal use
type UserWithPassword struct {
	*pb.User
	PasswordHash string
}
