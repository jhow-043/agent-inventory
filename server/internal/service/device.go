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

// ListDevices returns all devices, optionally filtered by hostname and OS.
func (s *DeviceService) ListDevices(ctx context.Context, hostname, osFilter string) (*dto.DeviceListResponse, error) {
	devices, err := s.deviceRepo.List(ctx, hostname, osFilter)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return &dto.DeviceListResponse{
		Devices: devices,
		Total:   len(devices),
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

	return &dto.DeviceDetailResponse{
		Device:            *device,
		Hardware:          hardware,
		Disks:             disks,
		NetworkInterfaces: network,
		InstalledSoftware: software,
	}, nil
}
