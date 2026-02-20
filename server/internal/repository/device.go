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
	Hostname     string
	OS           string
	Status       string // "online", "offline", "inactive", or "" (all active)
	DepartmentID string // UUID filter
	Sort         string // column name
	Order        string // "asc" or "desc"
	Page         int
	Limit        int
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

// Delete removes a device by ID. All related data is cascaded by the database.
func (r *DeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// List returns devices with filtering, sorting, and pagination.
// By default only active devices are returned; pass Status="inactive" to see inactive ones.
func (r *DeviceRepository) List(ctx context.Context, p ListParams) (*ListResult, error) {
	var where []string
	args := []interface{}{}
	argIdx := 1

	if p.Hostname != "" {
		where = append(where, fmt.Sprintf("d.hostname ILIKE $%d", argIdx))
		args = append(args, "%"+p.Hostname+"%")
		argIdx++
	}
	if p.OS != "" {
		where = append(where, fmt.Sprintf("(d.os_name ILIKE $%d OR d.os_version ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+p.OS+"%")
		argIdx++
	}
	if p.DepartmentID != "" {
		where = append(where, fmt.Sprintf("d.department_id = $%d", argIdx))
		args = append(args, p.DepartmentID)
		argIdx++
	}

	switch p.Status {
	case "online":
		where = append(where, "d.status = 'active'")
		where = append(where, "d.last_seen > NOW() - INTERVAL '1 hour'")
	case "offline":
		where = append(where, "d.status = 'active'")
		where = append(where, "d.last_seen <= NOW() - INTERVAL '1 hour'")
	case "inactive":
		where = append(where, "d.status = 'inactive'")
	default:
		// Empty status = all active devices
		where = append(where, "d.status = 'active'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	// Count total matching rows.
	countQuery := "SELECT COUNT(*) FROM devices d" + whereClause
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, fmt.Errorf("count devices: %w", err)
	}

	// Sorting — validate column.
	orderCol := "d.hostname"
	if col, ok := allowedSortColumns[p.Sort]; ok {
		orderCol = "d." + col
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

	dataQuery := fmt.Sprintf(`SELECT d.*, dep.name AS department_name
		FROM devices d
		LEFT JOIN departments dep ON dep.id = d.department_id
		%s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
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

// GetByID retrieves a single device by its primary key, including department name.
func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var device models.Device
	err := r.db.GetContext(ctx, &device, `SELECT d.*, dep.name AS department_name
		FROM devices d LEFT JOIN departments dep ON dep.id = d.department_id
		WHERE d.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetByHostname retrieves a single device by its hostname, including department name.
func (r *DeviceRepository) GetByHostname(ctx context.Context, hostname string) (*models.Device, error) {
	var device models.Device
	err := r.db.GetContext(ctx, &device, `SELECT d.*, dep.name AS department_name
		FROM devices d LEFT JOIN departments dep ON dep.id = d.department_id
		WHERE LOWER(d.hostname) = LOWER($1)`, hostname)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// UpdateStatus sets the status column of a device (active / inactive).
func (r *DeviceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE devices SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("update device status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// UpdateDepartment assigns a department (or NULL) to a device.
func (r *DeviceRepository) UpdateDepartment(ctx context.Context, id uuid.UUID, deptID *uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE devices SET department_id = $1 WHERE id = $2", deptID, id)
	if err != nil {
		return fmt.Errorf("update device department: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// GetHardwareHistory returns hardware change records for a device, newest first.
// If component is non-empty, only changes for that component are returned.
func (r *DeviceRepository) GetHardwareHistory(ctx context.Context, deviceID uuid.UUID, component string, limit, offset int) ([]models.HardwareHistory, int, error) {
	if limit <= 0 {
		limit = 50
	}

	args := []interface{}{deviceID}
	where := "WHERE device_id = $1 AND component IS NOT NULL"
	argIdx := 2
	if component != "" {
		where += fmt.Sprintf(" AND component = $%d", argIdx)
		args = append(args, component)
		argIdx++
	}

	// Count
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM hardware_history "+where, args...); err != nil {
		return nil, 0, fmt.Errorf("count hardware history: %w", err)
	}

	// Data
	query := fmt.Sprintf("SELECT * FROM hardware_history %s ORDER BY changed_at DESC LIMIT $%d OFFSET $%d", where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	var history []models.HardwareHistory
	if err := r.db.SelectContext(ctx, &history, query, args...); err != nil {
		return nil, 0, fmt.Errorf("get hardware history: %w", err)
	}
	if history == nil {
		history = []models.HardwareHistory{}
	}
	return history, total, nil
}

// ListForExport returns ALL devices matching the filters (no pagination) for CSV export.
func (r *DeviceRepository) ListForExport(ctx context.Context, p ListParams) ([]models.Device, error) {
	var where []string
	args := []interface{}{}
	argIdx := 1

	if p.Hostname != "" {
		where = append(where, fmt.Sprintf("d.hostname ILIKE $%d", argIdx))
		args = append(args, "%"+p.Hostname+"%")
		argIdx++
	}
	if p.OS != "" {
		where = append(where, fmt.Sprintf("(d.os_name ILIKE $%d OR d.os_version ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+p.OS+"%")
		argIdx++
	}
	if p.DepartmentID != "" {
		where = append(where, fmt.Sprintf("d.department_id = $%d", argIdx))
		args = append(args, p.DepartmentID)
		// argIdx not incremented as it's the last parameter
	}

	switch p.Status {
	case "online":
		where = append(where, "d.status = 'active'")
		where = append(where, "d.last_seen > NOW() - INTERVAL '1 hour'")
	case "offline":
		where = append(where, "d.status = 'active'")
		where = append(where, "d.last_seen <= NOW() - INTERVAL '1 hour'")
	case "inactive":
		where = append(where, "d.status = 'inactive'")
	default:
		where = append(where, "d.status = 'active'")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	orderCol := "d.hostname"
	if col, ok := allowedSortColumns[p.Sort]; ok {
		orderCol = "d." + col
	}
	orderDir := "ASC"
	if strings.EqualFold(p.Order, "desc") {
		orderDir = "DESC"
	}

	query := fmt.Sprintf(`SELECT d.*, dep.name AS department_name
		FROM devices d
		LEFT JOIN departments dep ON dep.id = d.department_id
		%s ORDER BY %s %s`, whereClause, orderCol, orderDir)

	var devices []models.Device
	if err := r.db.SelectContext(ctx, &devices, query, args...); err != nil {
		return nil, fmt.Errorf("list devices for export: %w", err)
	}
	if devices == nil {
		devices = []models.Device{}
	}
	return devices, nil
}

// BulkUpdateStatus sets the status column for multiple devices at once.
func (r *DeviceRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("UPDATE devices SET status = ? WHERE id IN (?)", status, ids)
	if err != nil {
		return 0, fmt.Errorf("build bulk status query: %w", err)
	}
	query = r.db.Rebind(query)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk update status: %w", err)
	}
	return res.RowsAffected()
}

// BulkUpdateDepartment sets the department for multiple devices at once.
func (r *DeviceRepository) BulkUpdateDepartment(ctx context.Context, ids []uuid.UUID, deptID *uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("UPDATE devices SET department_id = ? WHERE id IN (?)", deptID, ids)
	if err != nil {
		return 0, fmt.Errorf("build bulk dept query: %w", err)
	}
	query = r.db.Rebind(query)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk update department: %w", err)
	}
	return res.RowsAffected()
}

// BulkDelete deletes multiple devices by ID. Related data is cascaded.
func (r *DeviceRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("DELETE FROM devices WHERE id IN (?)", ids)
	if err != nil {
		return 0, fmt.Errorf("build bulk delete query: %w", err)
	}
	query = r.db.Rebind(query)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk delete devices: %w", err)
	}
	return res.RowsAffected()
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
