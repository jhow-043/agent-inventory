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
