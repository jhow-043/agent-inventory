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

	// OS distribution
	osRows, err := s.dashboardRepo.GetOSDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("get os distribution: %w", err)
	}
	osDist := make([]dto.ChartItem, len(osRows))
	for i, r := range osRows {
		osDist[i] = dto.ChartItem{Name: r.Name, Count: r.Count}
	}

	// Recent devices
	recentRows, err := s.dashboardRepo.GetRecentDevices(ctx, 5)
	if err != nil {
		return nil, fmt.Errorf("get recent devices: %w", err)
	}
	recent := make([]dto.RecentDevice, len(recentRows))
	for i, r := range recentRows {
		recent[i] = dto.RecentDevice{
			ID:       r.ID,
			Hostname: r.Hostname,
			OSName:   r.OSName,
			Status:   r.Status,
			LastSeen: r.LastSeen.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Top software
	swRows, err := s.dashboardRepo.GetTopSoftware(ctx, 8)
	if err != nil {
		return nil, fmt.Errorf("get top software: %w", err)
	}
	topSw := make([]dto.ChartItem, len(swRows))
	for i, r := range swRows {
		topSw[i] = dto.ChartItem{Name: r.Name, Count: r.Count}
	}

	return &dto.DashboardStatsResponse{
		Total:          total,
		Online:         online,
		Offline:        offline,
		Inactive:       inactive,
		OSDistribution: osDist,
		RecentDevices:  recent,
		TopSoftware:    topSw,
	}, nil
}
