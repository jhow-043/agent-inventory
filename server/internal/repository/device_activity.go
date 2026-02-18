package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// DeviceActivityRepository handles CRUD operations for device activity logs.
type DeviceActivityRepository struct {
	db *sqlx.DB
}

// NewDeviceActivityRepository creates a new DeviceActivityRepository.
func NewDeviceActivityRepository(db *sqlx.DB) *DeviceActivityRepository {
	return &DeviceActivityRepository{db: db}
}

// ActivityEntry represents a single activity to be inserted.
type ActivityEntry struct {
	DeviceID     uuid.UUID
	ActivityType string
	Description  string
	OldValue     *string
	NewValue     *string
	Metadata     *string
}

// InsertBatch inserts multiple activity entries in a single transaction.
func (r *DeviceActivityRepository) InsertBatch(ctx context.Context, tx *sqlx.Tx, entries []ActivityEntry) error {
	if len(entries) == 0 {
		return nil
	}

	stmt := `INSERT INTO device_activity_log (device_id, activity_type, description, old_value, new_value, metadata)
	         VALUES ($1, $2, $3, $4, $5, $6)`

	for _, e := range entries {
		if _, err := tx.ExecContext(ctx, stmt, e.DeviceID, e.ActivityType, e.Description, e.OldValue, e.NewValue, e.Metadata); err != nil {
			return fmt.Errorf("insert device activity: %w", err)
		}
	}
	return nil
}

// Insert inserts a single activity entry (standalone, no existing transaction).
func (r *DeviceActivityRepository) Insert(ctx context.Context, entry ActivityEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO device_activity_log (device_id, activity_type, description, old_value, new_value, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.DeviceID, entry.ActivityType, entry.Description, entry.OldValue, entry.NewValue, entry.Metadata)
	if err != nil {
		return fmt.Errorf("insert device activity: %w", err)
	}
	return nil
}

// ListByDevice returns activity logs for a device, ordered by most recent first.
func (r *DeviceActivityRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]models.DeviceActivityLog, int, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM device_activity_log WHERE device_id = $1", deviceID); err != nil {
		return nil, 0, fmt.Errorf("count device activities: %w", err)
	}

	var logs []models.DeviceActivityLog
	if err := r.db.SelectContext(ctx, &logs,
		`SELECT id, device_id, activity_type, description, old_value, new_value, metadata, detected_at
		 FROM device_activity_log
		 WHERE device_id = $1
		 ORDER BY detected_at DESC
		 LIMIT $2 OFFSET $3`,
		deviceID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list device activities: %w", err)
	}

	return logs, total, nil
}
