package repository

import (
	"context"
	"fmt"

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

// GetStats returns total and online device counts.
// A device is considered online if last_seen is within the last hour.
func (r *DashboardRepository) GetStats(ctx context.Context) (total int, online int, err error) {
	// Get total count.
	totalQuery := "SELECT COUNT(*) FROM devices"
	if err := r.db.GetContext(ctx, &total, totalQuery); err != nil {
		return 0, 0, fmt.Errorf("get total devices: %w", err)
	}

	// Get online count (last_seen within 1 hour).
	onlineQuery := "SELECT COUNT(*) FROM devices WHERE last_seen > NOW() - INTERVAL '1 hour'"
	if err := r.db.GetContext(ctx, &online, onlineQuery); err != nil {
		return 0, 0, fmt.Errorf("get online devices: %w", err)
	}

	return total, online, nil
}
