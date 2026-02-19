package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

// CleanupRepository handles data retention operations — purging old logs and history records.
type CleanupRepository struct {
	db *sqlx.DB
}

// NewCleanupRepository creates a new CleanupRepository.
func NewCleanupRepository(db *sqlx.DB) *CleanupRepository {
	return &CleanupRepository{db: db}
}

// CleanupResult holds the number of rows deleted for each table.
type CleanupResult struct {
	AuditLogs       int64
	ActivityLogs    int64
	HardwareHistory int64
}

// PurgeOldData removes records older than the specified retention days from log/history tables.
func (r *CleanupRepository) PurgeOldData(ctx context.Context, retentionDays int) (*CleanupResult, error) {
	result := &CleanupResult{}

	interval := fmt.Sprintf("%d days", retentionDays)

	// Purge audit_logs
	res, err := r.db.ExecContext(ctx,
		"DELETE FROM audit_logs WHERE created_at < NOW() - $1::interval", interval)
	if err != nil {
		return nil, fmt.Errorf("purge audit_logs: %w", err)
	}
	result.AuditLogs, _ = res.RowsAffected()

	// Purge device_activity_log
	res, err = r.db.ExecContext(ctx,
		"DELETE FROM device_activity_log WHERE detected_at < NOW() - $1::interval", interval)
	if err != nil {
		return nil, fmt.Errorf("purge device_activity_log: %w", err)
	}
	result.ActivityLogs, _ = res.RowsAffected()

	// Purge hardware_history (keep longer — snapshots are more valuable)
	res, err = r.db.ExecContext(ctx,
		"DELETE FROM hardware_history WHERE changed_at < NOW() - $1::interval", interval)
	if err != nil {
		return nil, fmt.Errorf("purge hardware_history: %w", err)
	}
	result.HardwareHistory, _ = res.RowsAffected()

	return result, nil
}

// CountRecords returns the current row counts for each log/history table.
func (r *CleanupRepository) CountRecords(ctx context.Context) (audit, activity, hardware int, err error) {
	if err = r.db.GetContext(ctx, &audit, "SELECT COUNT(*) FROM audit_logs"); err != nil {
		return
	}
	if err = r.db.GetContext(ctx, &activity, "SELECT COUNT(*) FROM device_activity_log"); err != nil {
		return
	}
	err = r.db.GetContext(ctx, &hardware, "SELECT COUNT(*) FROM hardware_history")
	return
}

// MarkInactiveDevices marks devices as inactive if they haven't been seen for the specified number of days.
func (r *CleanupRepository) MarkInactiveDevices(ctx context.Context, inactiveDays int) (int64, error) {
	interval := fmt.Sprintf("%d days", inactiveDays)
	res, err := r.db.ExecContext(ctx,
		"UPDATE devices SET status = 'inactive' WHERE status = 'active' AND last_seen < NOW() - $1::interval",
		interval)
	if err != nil {
		return 0, fmt.Errorf("mark inactive devices: %w", err)
	}
	return res.RowsAffected()
}

// VacuumAnalyze runs VACUUM ANALYZE on log tables to reclaim space.
func (r *CleanupRepository) VacuumAnalyze(ctx context.Context) error {
	tables := []string{"audit_logs", "device_activity_log", "hardware_history"}
	for _, table := range tables {
		if _, err := r.db.ExecContext(ctx, fmt.Sprintf("VACUUM ANALYZE %s", table)); err != nil {
			slog.Warn("vacuum analyze failed", "table", table, "error", err)
			// Don't return error — VACUUM failures are non-critical
		}
	}
	return nil
}
