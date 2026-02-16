package repository

import (
	"context"
	"fmt"
	"strings"

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

// ListParams holds filters, pagination, and sorting options for listing devices.
type ListParams struct {
	Hostname string
	OS       string
	Status   string // "online", "offline", or "" (all)
	Sort     string // column name
	Order    string // "asc" or "desc"
	Page     int
	Limit    int
}

// allowedSortColumns maps user-facing column names to SQL columns.
var allowedSortColumns = map[string]string{
	"hostname":  "hostname",
	"os":        "os_name",
	"last_seen": "last_seen",
	"status":    "last_seen",
}

// ListResult holds the paginated result from List.
type ListResult struct {
	Devices []models.Device
	Total   int
}

// List returns devices with filtering, sorting, and pagination.
func (r *DeviceRepository) List(ctx context.Context, p ListParams) (*ListResult, error) {
	var where []string
	args := []interface{}{}
	argIdx := 1

	if p.Hostname != "" {
		where = append(where, fmt.Sprintf("hostname ILIKE $%d", argIdx))
		args = append(args, "%"+p.Hostname+"%")
		argIdx++
	}
	if p.OS != "" {
		where = append(where, fmt.Sprintf("(os_name ILIKE $%d OR os_version ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+p.OS+"%")
		argIdx++
	}
	switch p.Status {
	case "online":
		where = append(where, "last_seen > NOW() - INTERVAL '1 hour'")
	case "offline":
		where = append(where, "last_seen <= NOW() - INTERVAL '1 hour'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	// Count total matching rows.
	countQuery := "SELECT COUNT(*) FROM devices" + whereClause
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, fmt.Errorf("count devices: %w", err)
	}

	// Sorting — validate column.
	orderCol := "hostname"
	if col, ok := allowedSortColumns[p.Sort]; ok {
		orderCol = col
	}
	orderDir := "ASC"
	if strings.EqualFold(p.Order, "desc") {
		orderDir = "DESC"
	}

	// Pagination defaults.
	limit := p.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	page := p.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	dataQuery := fmt.Sprintf("SELECT * FROM devices%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		whereClause, orderCol, orderDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	var devices []models.Device
	if err := r.db.SelectContext(ctx, &devices, dataQuery, args...); err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	if devices == nil {
		devices = []models.Device{}
	}

	return &ListResult{Devices: devices, Total: total}, nil
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
