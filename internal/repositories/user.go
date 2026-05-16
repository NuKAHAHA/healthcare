package repositories

import (
	"context"
	"fmt"
	"healthcare-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by email using parameterized query
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	// SECURITY: Using parameterized query to prevent SQL injection
	query := "SELECT id, email, first_name, last_name, role, password_hash, created_at, updated_at FROM users WHERE email = $1"
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID using parameterized query
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}

	// SECURITY: Using parameterized query to prevent SQL injection
	query := "SELECT id, email, first_name, last_name, role, password_hash, created_at, updated_at FROM users WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

// Create creates a new user using parameterized query
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		INSERT INTO users (email, first_name, last_name, role, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, email, first_name, last_name, role, password_hash, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Role,
		user.PasswordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByRole retrieves all users with a specific role
func (r *UserRepository) GetByRole(ctx context.Context, role string) ([]*models.User, error) {
	var users []*models.User

	// SECURITY: Using parameterized query to prevent SQL injection
	query := "SELECT id, email, first_name, last_name, role, password_hash, created_at, updated_at FROM users WHERE role = $1 ORDER BY created_at DESC"
	rows, err := r.db.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
