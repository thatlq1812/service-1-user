package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "service-1-user/proto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		// No row
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailDuplicate
		}
		return nil, fmt.Errorf("Insert user failed: %w", err)
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil
}

// GetByEmailWithPassword implement method to get user by email with password hash
func (r *userPostgresRepo) GetByEmailWithPassword(ctx context.Context, email string) (*UserWithPassword, error) {
	query := `
	SELECT id, name, email, password_hash, created_at
	FROM users
	WHERE email = $1
	`

	var user pb.User
	var createdAt time.Time
	var passwordHash string

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&passwordHash,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("Query user by email failed: %w", err)
	}

	// Convert time.Time to string
	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &UserWithPassword{
		User:         &user,
		PasswordHash: passwordHash,
	}, nil
}

// CreateWithPassword implement method to create new user with password
func (r *userPostgresRepo) CreateWithPassword(ctx context.Context, name, email, passwordHash string) (*pb.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`

	var user pb.User
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query, name, email, passwordHash).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&createdAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailDuplicate
		}

		return nil, fmt.Errorf("Insert user with password failed: %w", err)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailDuplicate
		}
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
		return ErrUserNotFound
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
