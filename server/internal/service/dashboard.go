package service

import (
	"context"
	"fmt"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
)

// DashboardService handles dashboard statistics business logic.
type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: repo}
}

// GetStats returns aggregated dashboard statistics.
func (s *DashboardService) GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	total, online, inactive, err := s.dashboardRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}

	offline := total - online

	return &dto.DashboardStatsResponse{
		Total:    total,
		Online:   online,
		Offline:  offline,
		Inactive: inactive,
	}, nil
}
