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
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
}

// UserResponse is a single user returned by the API (no password hash).
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
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
}

// DeviceDetailResponse is returned by GET /api/v1/devices/:id.
type DeviceDetailResponse struct {
	Device            models.Device              `json:"device"`
	Hardware          *models.Hardware           `json:"hardware"`
	Disks             []models.Disk              `json:"disks"`
	NetworkInterfaces []models.NetworkInterface  `json:"network_interfaces"`
	InstalledSoftware []models.InstalledSoftware `json:"installed_software"`
	RemoteTools       []models.RemoteTool        `json:"remote_tools"`
}
