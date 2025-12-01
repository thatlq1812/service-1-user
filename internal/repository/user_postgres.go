package repository

import (
	"context"
	"fmt"
	"time"

	pb "service-1-user/proto"

	"github.com/jackc/pgx/v5/pgxpool"
)

// userPostgresRepo implement User repository with PostgresSQL
type userPostgresRepo struct {
	db *pgxpool.Pool
}

// NewUserPostgresRepository create new instance
func NewUserPostgresRepository(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepo{db: db}
}

// GetByID implement method with user by ID
func (r *userPostgresRepo) GetByID(ctx context.Context, id int32) (*pb.User, error) {
	query := `
	SELECT id, name, email, created_at
	FROM users
	WHERE id = $1
	`

	var user pb.User
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Query user failed: %w", err)
	}

	// Convert time.Time to string
	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil
}

// Create implement method to create new user
func (r *userPostgresRepo) Create(ctx context.Context, name, email string) (*pb.User, error) {
	query := `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, name, email, created_at
	`

	var user pb.User
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query, name, email).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Insert user failed: %w", err)
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil

}

// Update implement method for update user
func (r *userPostgresRepo) Update(ctx context.Context, id int32, name, email string) (*pb.User, error) {
	query := `
		UPDATE users
		SET name = $1, email = $2
		WHERE id = $3
		RETURNING id, name, email, created_at
	`

	var user pb.User
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query, name, email, id).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Update user failed: %w", err)
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil

}

// Delete implement method to delete user by ID
func (r *userPostgresRepo) Delete(ctx context.Context, id int32) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Delete user failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("No user found with ID %d", id)
	}

	return nil
}

// List implement method to get list of all users
func (r *userPostgresRepo) List(ctx context.Context, limit, offset int32) ([]*pb.User, int32, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("Query users failed: %w", err)
	}
	defer rows.Close()

	var users []*pb.User

	for rows.Next() {
		var user pb.User
		var createdAt time.Time
		err := rows.Scan(
			&user.Id,
			&user.Name,
			&user.Email,
			&createdAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("Scan user failed: %w", err)
		}

		user.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, &user)
	}

	// Query for count total users
	countQuery := `SELECT COUNT(*) FROM users`
	var total int32
	err = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("Count users failed: %w", err)
	}

	return users, total, nil
}
