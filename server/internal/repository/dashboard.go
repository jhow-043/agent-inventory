package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DashboardRepository handles dashboard statistics queries.
type DashboardRepository struct {
	db *sqlx.DB
}

// NewDashboardRepository creates a new DashboardRepository.
func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetStats returns total, online, and inactive device counts.
// Only active devices count toward total/online/offline. Inactive is separate.
func (r *DashboardRepository) GetStats(ctx context.Context) (total int, online int, inactive int, err error) {
	// Get active device count.
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM devices WHERE status = 'active'"); err != nil {
		return 0, 0, 0, fmt.Errorf("get total devices: %w", err)
	}

	// Get online count (active + last_seen within 1 hour).
	if err := r.db.GetContext(ctx, &online,
		"SELECT COUNT(*) FROM devices WHERE status = 'active' AND last_seen > NOW() - INTERVAL '1 hour'"); err != nil {
		return 0, 0, 0, fmt.Errorf("get online devices: %w", err)
	}

	// Get inactive count.
	if err := r.db.GetContext(ctx, &inactive, "SELECT COUNT(*) FROM devices WHERE status = 'inactive'"); err != nil {
		return 0, 0, 0, fmt.Errorf("get inactive devices: %w", err)
	}

	return total, online, inactive, nil
}

// OSCount holds the OS name and its device count.
type OSCount struct {
	Name  string `db:"os_name"`
	Count int    `db:"count"`
}

// GetOSDistribution returns counts of devices grouped by OS name.
func (r *DashboardRepository) GetOSDistribution(ctx context.Context) ([]OSCount, error) {
	var result []OSCount
	err := r.db.SelectContext(ctx, &result,
		`SELECT COALESCE(os_name, 'Unknown') AS os_name, COUNT(*) AS count
		 FROM devices WHERE status != 'inactive'
		 GROUP BY os_name ORDER BY count DESC LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("get os distribution: %w", err)
	}
	return result, nil
}

// RecentDeviceRow is a minimal device representation for recent activity.
type RecentDeviceRow struct {
	ID       uuid.UUID `db:"id"`
	Hostname string    `db:"hostname"`
	OSName   string    `db:"os_name"`
	Status   string    `db:"status"`
	LastSeen time.Time `db:"last_seen"`
}

// GetRecentDevices returns the most recently seen devices.
func (r *DashboardRepository) GetRecentDevices(ctx context.Context, limit int) ([]RecentDeviceRow, error) {
	var result []RecentDeviceRow
	err := r.db.SelectContext(ctx, &result,
		`SELECT id, hostname, COALESCE(os_name, '') AS os_name, status, last_seen
		 FROM devices ORDER BY last_seen DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent devices: %w", err)
	}
	return result, nil
}

// SoftwareCount holds a software name and its install count.
type SoftwareCount struct {
	Name  string `db:"name"`
	Count int    `db:"count"`
}

// GetTopSoftware returns the most commonly installed software across all devices.
func (r *DashboardRepository) GetTopSoftware(ctx context.Context, limit int) ([]SoftwareCount, error) {
	var result []SoftwareCount
	err := r.db.SelectContext(ctx, &result,
		`SELECT name, COUNT(DISTINCT device_id) AS count
		 FROM installed_software
		 GROUP BY name ORDER BY count DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("get top software: %w", err)
	}
	return result, nil
}
