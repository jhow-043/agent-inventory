package service

import (
	"context"
	"log/slog"
	"time"

	"inventario/server/internal/repository"
)

// CleanupService handles scheduled data retention and housekeeping tasks.
type CleanupService struct {
	repo          *repository.CleanupRepository
	retentionDays int
	inactiveDays  int
	interval      time.Duration
	stopCh        chan struct{}
}

// NewCleanupService creates a new CleanupService.
// retentionDays: records older than this are purged (default 90).
// inactiveDays: devices not seen for this many days are marked inactive (default 30).
// interval: how often the cleanup runs (default 24h).
func NewCleanupService(repo *repository.CleanupRepository, retentionDays, inactiveDays int, interval time.Duration) *CleanupService {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	if inactiveDays <= 0 {
		inactiveDays = 30
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &CleanupService{
		repo:          repo,
		retentionDays: retentionDays,
		inactiveDays:  inactiveDays,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the periodic cleanup loop in a background goroutine.
// Call Stop() to terminate it gracefully.
func (s *CleanupService) Start() {
	// Run once immediately at startup
	go func() {
		s.runCleanup()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runCleanup()
			case <-s.stopCh:
				slog.Info("cleanup service stopped")
				return
			}
		}
	}()

	slog.Info("cleanup service started",
		"retention_days", s.retentionDays,
		"inactive_days", s.inactiveDays,
		"interval", s.interval.String(),
	)
}

// Stop signals the cleanup goroutine to terminate.
func (s *CleanupService) Stop() {
	close(s.stopCh)
}

func (s *CleanupService) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Info("running scheduled data cleanup")

	// 1. Count before
	auditBefore, activityBefore, hwBefore, err := s.repo.CountRecords(ctx)
	if err != nil {
		slog.Error("cleanup: failed to count records", "error", err)
	}

	// 2. Purge old data
	result, err := s.repo.PurgeOldData(ctx, s.retentionDays)
	if err != nil {
		slog.Error("cleanup: failed to purge old data", "error", err)
		return
	}

	// 3. Mark inactive devices
	inactiveCount, err := s.repo.MarkInactiveDevices(ctx, s.inactiveDays)
	if err != nil {
		slog.Error("cleanup: failed to mark inactive devices", "error", err)
	}

	// 4. Log results
	totalPurged := result.AuditLogs + result.ActivityLogs + result.HardwareHistory
	if totalPurged > 0 || inactiveCount > 0 {
		slog.Info("cleanup completed",
			"audit_logs_purged", result.AuditLogs,
			"activity_logs_purged", result.ActivityLogs,
			"hardware_history_purged", result.HardwareHistory,
			"devices_marked_inactive", inactiveCount,
			"audit_before", auditBefore,
			"activity_before", activityBefore,
			"hardware_before", hwBefore,
		)
	} else {
		slog.Debug("cleanup completed — nothing to purge")
	}
}
