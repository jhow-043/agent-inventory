package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// DeviceRepository handles read queries for devices and their related data.
type DeviceRepository struct {
	db *sqlx.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// List returns all devices, optionally filtered by hostname and/or OS.
func (r *DeviceRepository) List(ctx context.Context, hostname, osFilter string) ([]models.Device, error) {
	query := "SELECT * FROM devices WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if hostname != "" {
		query += fmt.Sprintf(" AND hostname ILIKE $%d", argIdx)
		args = append(args, "%"+hostname+"%")
		argIdx++
	}
	if osFilter != "" {
		query += fmt.Sprintf(" AND (os_name ILIKE $%d OR os_version ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+osFilter+"%")
		argIdx++
	}

	query += " ORDER BY hostname"

	var devices []models.Device
	err := r.db.SelectContext(ctx, &devices, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	if devices == nil {
		devices = []models.Device{}
	}
	return devices, nil
}

// GetByID retrieves a single device by its primary key.
func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var device models.Device
	err := r.db.GetContext(ctx, &device, "SELECT * FROM devices WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetBySerialNumber retrieves a device by its serial number.
func (r *DeviceRepository) GetBySerialNumber(ctx context.Context, sn string) (*models.Device, error) {
	var device models.Device
	err := r.db.GetContext(ctx, &device, "SELECT * FROM devices WHERE serial_number = $1", sn)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetHardware retrieves the hardware record for a device.
func (r *DeviceRepository) GetHardware(ctx context.Context, deviceID uuid.UUID) (*models.Hardware, error) {
	var hw models.Hardware
	err := r.db.GetContext(ctx, &hw, "SELECT * FROM hardware WHERE device_id = $1", deviceID)
	if err != nil {
		return nil, err
	}
	return &hw, nil
}

// GetDisks retrieves all disk records for a device.
func (r *DeviceRepository) GetDisks(ctx context.Context, deviceID uuid.UUID) ([]models.Disk, error) {
	var disks []models.Disk
	err := r.db.SelectContext(ctx, &disks, "SELECT * FROM disks WHERE device_id = $1 ORDER BY model", deviceID)
	if err != nil {
		return nil, err
	}
	if disks == nil {
		disks = []models.Disk{}
	}
	return disks, nil
}

// GetNetworkInterfaces retrieves all network interface records for a device.
func (r *DeviceRepository) GetNetworkInterfaces(ctx context.Context, deviceID uuid.UUID) ([]models.NetworkInterface, error) {
	var nics []models.NetworkInterface
	err := r.db.SelectContext(ctx, &nics, "SELECT * FROM network_interfaces WHERE device_id = $1 ORDER BY name", deviceID)
	if err != nil {
		return nil, err
	}
	if nics == nil {
		nics = []models.NetworkInterface{}
	}
	return nics, nil
}

// GetInstalledSoftware retrieves all installed software records for a device.
func (r *DeviceRepository) GetInstalledSoftware(ctx context.Context, deviceID uuid.UUID) ([]models.InstalledSoftware, error) {
	var sw []models.InstalledSoftware
	err := r.db.SelectContext(ctx, &sw, "SELECT * FROM installed_software WHERE device_id = $1 ORDER BY name", deviceID)
	if err != nil {
		return nil, err
	}
	if sw == nil {
		sw = []models.InstalledSoftware{}
	}
	return sw, nil
}

// GetRemoteTools retrieves all remote tool records for a device.
func (r *DeviceRepository) GetRemoteTools(ctx context.Context, deviceID uuid.UUID) ([]models.RemoteTool, error) {
	var tools []models.RemoteTool
	err := r.db.SelectContext(ctx, &tools, "SELECT * FROM remote_tools WHERE device_id = $1 ORDER BY tool_name", deviceID)
	if err != nil {
		return nil, err
	}
	if tools == nil {
		tools = []models.RemoteTool{}
	}
	return tools, nil
}
