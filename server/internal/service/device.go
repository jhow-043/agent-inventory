package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
	"inventario/shared/models"
)

// DeviceService handles device listing and detail queries.
type DeviceService struct {
	deviceRepo *repository.DeviceRepository
}

// NewDeviceService creates a new DeviceService.
func NewDeviceService(repo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{deviceRepo: repo}
}

// ListDevices returns devices with pagination, filtering, and sorting.
func (s *DeviceService) ListDevices(ctx context.Context, p repository.ListParams) (*dto.DeviceListResponse, error) {
	result, err := s.deviceRepo.List(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return &dto.DeviceListResponse{
		Devices: result.Devices,
		Total:   result.Total,
		Page:    p.Page,
		Limit:   p.Limit,
	}, nil
}

// GetDeviceDetail returns a device with all its related data.
func (s *DeviceService) GetDeviceDetail(ctx context.Context, id uuid.UUID) (*dto.DeviceDetailResponse, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("get device: %w", err)
	}

	// Fetch related data — errors for missing records are acceptable (empty slices).
	hardware, _ := s.deviceRepo.GetHardware(ctx, id)
	disks, err := s.deviceRepo.GetDisks(ctx, id)
	if err != nil {
		disks = []models.Disk{}
	}
	network, err := s.deviceRepo.GetNetworkInterfaces(ctx, id)
	if err != nil {
		network = []models.NetworkInterface{}
	}
	software, err := s.deviceRepo.GetInstalledSoftware(ctx, id)
	if err != nil {
		software = []models.InstalledSoftware{}
	}
	remoteTools, err := s.deviceRepo.GetRemoteTools(ctx, id)
	if err != nil {
		remoteTools = []models.RemoteTool{}
	}
	hwHistory, err := s.deviceRepo.GetHardwareHistory(ctx, id)
	if err != nil {
		hwHistory = []models.HardwareHistory{}
	}

	return &dto.DeviceDetailResponse{
		Device:            *device,
		Hardware:          hardware,
		Disks:             disks,
		NetworkInterfaces: network,
		InstalledSoftware: software,
		RemoteTools:       remoteTools,
		HardwareHistory:   hwHistory,
	}, nil
}

// UpdateStatus changes the status of a device (active / inactive).
func (s *DeviceService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.deviceRepo.UpdateStatus(ctx, id, status)
}

// UpdateDepartment changes the department assignment for a device.
func (s *DeviceService) UpdateDepartment(ctx context.Context, id uuid.UUID, deptID *uuid.UUID) error {
	return s.deviceRepo.UpdateDepartment(ctx, id, deptID)
}

// ListForExport returns all devices matching the filters (no pagination) for CSV export.
func (s *DeviceService) ListForExport(ctx context.Context, p repository.ListParams) ([]models.Device, error) {
	return s.deviceRepo.ListForExport(ctx, p)
}
