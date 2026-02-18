package dto

import (
	"inventario/shared/models"

	"github.com/google/uuid"
)

// ErrorResponse is returned for any API error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse is a generic success message.
type MessageResponse struct {
	Message string `json:"message"`
}

// MeResponse is returned by GET /api/v1/auth/me.
type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// HealthResponse is returned by the liveness probe.
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadyResponse is returned by the readiness probe.
type ReadyResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// DashboardStatsResponse is returned by GET /api/v1/dashboard/stats.
type DashboardStatsResponse struct {
	Total          int            `json:"total"`
	Online         int            `json:"online"`
	Offline        int            `json:"offline"`
	Inactive       int            `json:"inactive"`
	OSDistribution []ChartItem    `json:"os_distribution"`
	RecentDevices  []RecentDevice `json:"recent_devices"`
	TopSoftware    []ChartItem    `json:"top_software"`
}

// ChartItem is a generic name/value pair for charts.
type ChartItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// RecentDevice is a minimal device representation for the dashboard.
type RecentDevice struct {
	ID       uuid.UUID `json:"id"`
	Hostname string    `json:"hostname"`
	OSName   string    `json:"os_name"`
	Status   string    `json:"status"`
	LastSeen string    `json:"last_seen"`
}

// UserResponse is a single user returned by the API (no password hash).
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"created_at"`
}

// UserListResponse is returned by GET /api/v1/users.
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// EnrollResponse is returned after a successful device enrollment.
type EnrollResponse struct {
	DeviceID uuid.UUID `json:"device_id"`
	Token    string    `json:"token"`
}

// DeviceListResponse is returned by GET /api/v1/devices.
type DeviceListResponse struct {
	Devices []models.Device `json:"devices"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
}

// DeviceDetailResponse is returned by GET /api/v1/devices/:id.
type DeviceDetailResponse struct {
	Device            models.Device              `json:"device"`
	Hardware          *models.Hardware           `json:"hardware"`
	Disks             []models.Disk              `json:"disks"`
	NetworkInterfaces []models.NetworkInterface  `json:"network_interfaces"`
	InstalledSoftware []models.InstalledSoftware `json:"installed_software"`
	RemoteTools       []models.RemoteTool        `json:"remote_tools"`
	HardwareHistory   []models.HardwareHistory   `json:"hardware_history"`
}

// DepartmentResponse is returned for department CRUD operations.
type DepartmentResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
}

// DepartmentListResponse is returned by GET /api/v1/departments.
type DepartmentListResponse struct {
	Departments []models.Department `json:"departments"`
	Total       int                 `json:"total"`
}

// AuditLogResponse is a single audit log entry.
type AuditLogResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id"`
	Username     string     `json:"username"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Details      string     `json:"details,omitempty"`
	IPAddress    string     `json:"ip_address,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	CreatedAt    string     `json:"created_at"`
}

// AuditLogListResponse is returned by GET /api/v1/audit-logs.
type AuditLogListResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int                `json:"total"`
}
