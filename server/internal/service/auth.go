package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"inventario/server/internal/middleware"
	"inventario/server/internal/repository"
	"inventario/shared/dto"
	"inventario/shared/models"
)

// AuthService handles enrollment, login, and user management.
type AuthService struct {
	db        *sqlx.DB
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *sqlx.DB, userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, jwtSecret string) *AuthService {
	return &AuthService{db: db, userRepo: userRepo, tokenRepo: tokenRepo, jwtSecret: jwtSecret}
}

// Enroll registers a new agent or re-enrolls an existing one.
// It creates the device if it does not exist (by serial_number), then generates a fresh token.
func (s *AuthService) Enroll(ctx context.Context, req *dto.EnrollRequest) (*dto.EnrollResponse, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Check if device already exists by serial number.
	var device models.Device
	err = tx.GetContext(ctx, &device, "SELECT * FROM devices WHERE serial_number = $1", req.SerialNumber)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// New device — insert.
		device.ID = uuid.New()
		device.Hostname = req.Hostname
		device.SerialNumber = req.SerialNumber
		if _, err = tx.ExecContext(ctx,
			"INSERT INTO devices (id, hostname, serial_number, last_seen) VALUES ($1, $2, $3, NOW())",
			device.ID, device.Hostname, device.SerialNumber,
		); err != nil {
			return nil, fmt.Errorf("insert device: %w", err)
		}
		slog.Info("new device created via enrollment", "device_id", device.ID, "hostname", req.Hostname)
	case err != nil:
		return nil, fmt.Errorf("query device: %w", err)
	default:
		// Existing device — update hostname and last_seen.
		if _, err = tx.ExecContext(ctx,
			"UPDATE devices SET hostname = $1, last_seen = NOW(), updated_at = NOW() WHERE id = $2",
			req.Hostname, device.ID,
		); err != nil {
			return nil, fmt.Errorf("update device: %w", err)
		}
		slog.Info("device re-enrolled", "device_id", device.ID, "hostname", req.Hostname)
	}

	// Remove previous token (if any).
	if _, err = tx.ExecContext(ctx, "DELETE FROM device_tokens WHERE device_id = $1", device.ID); err != nil {
		return nil, fmt.Errorf("delete old token: %w", err)
	}

	// Generate a new token.
	rawToken := uuid.New().String()
	tokenHash := middleware.SHA256Hex(rawToken)

	if _, err = tx.ExecContext(ctx,
		"INSERT INTO device_tokens (id, device_id, token_hash) VALUES ($1, $2, $3)",
		uuid.New(), device.ID, tokenHash,
	); err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &dto.EnrollResponse{
		DeviceID: device.ID,
		Token:    rawToken,
	}, nil
}

// Login authenticates a dashboard user and returns a signed JWT string.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

// CreateUser creates a new dashboard user with a bcrypt-hashed password.
func (s *AuthService) CreateUser(ctx context.Context, username, password, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Default to viewer if role is empty
	if role == "" {
		role = "viewer"
	}

	// Validate role
	if role != "admin" && role != "viewer" {
		return fmt.Errorf("invalid role: must be admin or viewer")
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}

	return s.userRepo.Create(ctx, user)
}

// ListUsers returns all dashboard users.
func (s *AuthService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.userRepo.List(ctx)
}

// DeleteUser deletes a dashboard user by ID.
// It prevents a user from deleting themselves.
func (s *AuthService) DeleteUser(ctx context.Context, requestingUserID, targetUserID uuid.UUID) error {
	if requestingUserID == targetUserID {
		return fmt.Errorf("cannot delete your own account")
	}
	return s.userRepo.Delete(ctx, targetUserID)
}
