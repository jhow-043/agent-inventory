// Package repository implements the data access layer using sqlx.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// TokenRepository handles device token persistence.
type TokenRepository struct {
	db *sqlx.DB
}

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(db *sqlx.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// GetByHash retrieves a device token by its SHA-256 hash.
func (r *TokenRepository) GetByHash(ctx context.Context, hash string) (*models.DeviceToken, error) {
	var token models.DeviceToken
	err := r.db.GetContext(ctx, &token, "SELECT * FROM device_tokens WHERE token_hash = $1", hash)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Create inserts a new device token.
func (r *TokenRepository) Create(ctx context.Context, token *models.DeviceToken) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO device_tokens (id, device_id, token_hash) VALUES ($1, $2, $3)",
		token.ID, token.DeviceID, token.TokenHash)
	return err
}

// DeleteByDeviceID removes all tokens belonging to a specific device.
func (r *TokenRepository) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM device_tokens WHERE device_id = $1", deviceID)
	return err
}
