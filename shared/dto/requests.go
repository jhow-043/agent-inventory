// Package dto defines the data transfer objects for API requests.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// EnrollRequest is sent by the agent to register with the API.
type EnrollRequest struct {
	Hostname     string `json:"hostname" binding:"required"`
	SerialNumber string `json:"serial_number" binding:"required"`
}

// InventoryRequest is the full inventory payload sent by the agent.
type InventoryRequest struct {
	Hostname      string           `json:"hostname" binding:"required"`
	SerialNumber  string           `json:"serial_number" binding:"required"`
	OSName        string           `json:"os_name"`
	OSVersion     string           `json:"os_version"`
	OSBuild       string           `json:"os_build"`
	OSArch        string           `json:"os_arch"`
	LastBootTime  *time.Time       `json:"last_boot_time,omitempty"`
	LoggedInUser  string           `json:"logged_in_user"`
	AgentVersion  string           `json:"agent_version"`
	LicenseStatus string           `json:"license_status"`
	Hardware      HardwareData     `json:"hardware"`
	Disks         []DiskData       `json:"disks"`
	Network       []NetworkData    `json:"network_interfaces"`
	Software      []SoftwareData   `json:"installed_software"`
	RemoteTools   []RemoteToolData `json:"remote_tools"`
}

// HardwareData contains CPU, RAM, motherboard, and BIOS info.
type HardwareData struct {
	CPUModel                string `json:"cpu_model"`
	CPUCores                int    `json:"cpu_cores"`
	CPUThreads              int    `json:"cpu_threads"`
	RAMTotalBytes           int64  `json:"ram_total_bytes"`
	MotherboardManufacturer string `json:"motherboard_manufacturer"`
	MotherboardProduct      string `json:"motherboard_product"`
	MotherboardSerial       string `json:"motherboard_serial"`
	BIOSVendor              string `json:"bios_vendor"`
	BIOSVersion             string `json:"bios_version"`
}

// DiskData contains disk drive information.
type DiskData struct {
	Model              string `json:"model"`
	SizeBytes          int64  `json:"size_bytes"`
	MediaType          string `json:"media_type"`
	SerialNumber       string `json:"serial_number"`
	InterfaceType      string `json:"interface_type"`
	DriveLetter        string `json:"drive_letter"`
	PartitionSizeBytes int64  `json:"partition_size_bytes"`
	FreeSpaceBytes     int64  `json:"free_space_bytes"`
}

// NetworkData contains network interface information.
type NetworkData struct {
	Name        string `json:"name"`
	MACAddress  string `json:"mac_address"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
	SpeedMbps   *int   `json:"speed_mbps,omitempty"`
	IsPhysical  bool   `json:"is_physical"`
}

// SoftwareData contains installed software information.
type SoftwareData struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Vendor      string `json:"vendor"`
	InstallDate string `json:"install_date"`
}

// LoginRequest is used to authenticate dashboard users.
type RemoteToolData struct {
	ToolName string `json:"tool_name"`
	RemoteID string `json:"remote_id"`
	Version  string `json:"version"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,max=100"`
	Password string `json:"password" binding:"required,max=200"`
}

// CreateUserRequest is used to create a new dashboard user.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Name     string `json:"name" binding:"required,max=255"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=admin viewer"`
}

// UpdateUserRequest is used to update a dashboard user's info.
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=100"`
	Name     string `json:"name" binding:"omitempty,max=255"`
	Password string `json:"password" binding:"omitempty,min=8,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=admin viewer"`
}

// UpdateDeviceStatusRequest is used to change a device's lifecycle status.
type UpdateDeviceStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateDeviceDepartmentRequest is used to assign a device to a department.
type UpdateDeviceDepartmentRequest struct {
	DepartmentID *uuid.UUID `json:"department_id"`
}

// BulkDeviceStatusRequest is used to change the status of multiple devices at once.
type BulkDeviceStatusRequest struct {
	DeviceIDs []uuid.UUID `json:"device_ids" binding:"required,min=1,max=100"`
	Status    string      `json:"status" binding:"required,oneof=active inactive"`
}

// BulkDeviceDepartmentRequest is used to assign multiple devices to a department.
type BulkDeviceDepartmentRequest struct {
	DeviceIDs    []uuid.UUID `json:"device_ids" binding:"required,min=1,max=100"`
	DepartmentID *uuid.UUID  `json:"department_id"`
}

// BulkDeviceDeleteRequest is used to delete multiple devices at once.
type BulkDeviceDeleteRequest struct {
	DeviceIDs []uuid.UUID `json:"device_ids" binding:"required,min=1,max=100"`
}

// BulkActionResponse returns the count of affected devices.
type BulkActionResponse struct {
	Affected int    `json:"affected"`
	Message  string `json:"message"`
}

// CreateDepartmentRequest is used to create a new department.
type CreateDepartmentRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// UpdateDepartmentRequest is used to rename a department.
type UpdateDepartmentRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}
