package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// UserRepository handles dashboard user persistence.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername retrieves a user by their username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)",
		user.ID, user.Username, user.PasswordHash)
	return err
}

// List returns all users ordered by creation date (newest first).
func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = []models.User{}
	}
	return users, nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
