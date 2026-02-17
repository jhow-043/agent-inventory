package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// AuditLogRepository handles audit log persistence.
type AuditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create inserts a new audit log entry.
func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs 
		(id, user_id, username, action, resource_type, resource_id, details, ip_address, user_agent, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`,
		log.ID, log.UserID, log.Username, log.Action, log.ResourceType, log.ResourceID, log.Details, log.IPAddress, log.UserAgent,
	)
	return err
}

// List returns audit logs with filtering and pagination.
func (r *AuditLogRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.AuditLog, int, error) {
	query := "SELECT * FROM audit_logs WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		query += " AND user_id = $" + string(rune(argIndex+'0'))
		countQuery += " AND user_id = $" + string(rune(argIndex+'0'))
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"].(string); ok {
		query += " AND action = $" + string(rune(argIndex+'0'))
		countQuery += " AND action = $" + string(rune(argIndex+'0'))
		args = append(args, action)
		argIndex++
	}

	if resourceType, ok := filters["resource_type"].(string); ok {
		query += " AND resource_type = $" + string(rune(argIndex+'0'))
		countQuery += " AND resource_type = $" + string(rune(argIndex+'0'))
		args = append(args, resourceType)
		argIndex++
	}

	if resourceID, ok := filters["resource_id"].(uuid.UUID); ok {
		query += " AND resource_id = $" + string(rune(argIndex+'0'))
		countQuery += " AND resource_id = $" + string(rune(argIndex+'0'))
		args = append(args, resourceID)
		argIndex++
	}

	// Get total count
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Order by newest first
	query += " ORDER BY created_at DESC"

	// Apply pagination
	query += " LIMIT $" + string(rune(argIndex+'0')) + " OFFSET $" + string(rune(argIndex+'1'))
	args = append(args, limit, offset)

	var logs []models.AuditLog
	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return nil, 0, err
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, total, nil
}

// GetByResourceID returns all audit logs for a specific resource.
func (r *AuditLogRepository) GetByResourceID(ctx context.Context, resourceType string, resourceID uuid.UUID, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.SelectContext(ctx, &logs,
		`SELECT * FROM audit_logs 
		WHERE resource_type = $1 AND resource_id = $2 
		ORDER BY created_at DESC LIMIT $3`,
		resourceType, resourceID, limit,
	)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	return logs, nil
}

// CreateWithDetails is a helper that marshals details to JSON and creates the log.
func (r *AuditLogRepository) CreateWithDetails(ctx context.Context, userID *uuid.UUID, username, action, resourceType string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	var detailsJSON string
	if details != nil {
		b, err := json.Marshal(details)
		if err != nil {
			return err
		}
		detailsJSON = string(b)
	}

	log := &models.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      detailsJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return r.Create(ctx, log)
}
