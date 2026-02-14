// Package dto defines the data transfer objects for API requests.
package dto

import "time"

// EnrollRequest is sent by the agent to register with the API.
type EnrollRequest struct {
	Hostname     string `json:"hostname" binding:"required"`
	SerialNumber string `json:"serial_number" binding:"required"`
}

// InventoryRequest is the full inventory payload sent by the agent.
type InventoryRequest struct {
	Hostname      string           `json:"hostname" binding:"required"`
	SerialNumber  string           `json:"serial_number" binding:"required"`
	OSName        string           `json:"os_name" binding:"required"`
	OSVersion     string           `json:"os_version" binding:"required"`
	OSBuild       string           `json:"os_build"`
	OSArch        string           `json:"os_arch"`
	LastBootTime  *time.Time       `json:"last_boot_time,omitempty"`
	LoggedInUser  string           `json:"logged_in_user"`
	AgentVersion  string           `json:"agent_version"`
	LicenseStatus string           `json:"license_status"`
	Hardware      HardwareData     `json:"hardware" binding:"required"`
	Disks         []DiskData       `json:"disks"`
	Network       []NetworkData    `json:"network_interfaces"`
	Software      []SoftwareData   `json:"installed_software"`
	RemoteTools   []RemoteToolData `json:"remote_tools"`
}

// HardwareData contains CPU, RAM, motherboard, and BIOS info.
type HardwareData struct {
	CPUModel                string `json:"cpu_model" binding:"required"`
	CPUCores                int    `json:"cpu_cores" binding:"required"`
	CPUThreads              int    `json:"cpu_threads"`
	RAMTotalBytes           int64  `json:"ram_total_bytes" binding:"required"`
	MotherboardManufacturer string `json:"motherboard_manufacturer"`
	MotherboardProduct      string `json:"motherboard_product"`
	MotherboardSerial       string `json:"motherboard_serial"`
	BIOSVendor              string `json:"bios_vendor"`
	BIOSVersion             string `json:"bios_version"`
}

// DiskData contains disk drive information.
type DiskData struct {
	Model         string `json:"model" binding:"required"`
	SizeBytes     int64  `json:"size_bytes" binding:"required"`
	MediaType     string `json:"media_type"`
	SerialNumber  string `json:"serial_number"`
	InterfaceType string `json:"interface_type"`
}

// NetworkData contains network interface information.
type NetworkData struct {
	Name        string `json:"name" binding:"required"`
	MACAddress  string `json:"mac_address"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
	SpeedMbps   *int   `json:"speed_mbps,omitempty"`
	IsPhysical  bool   `json:"is_physical"`
}

// SoftwareData contains installed software information.
type SoftwareData struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version"`
	Vendor      string `json:"vendor"`
	InstallDate string `json:"install_date"`
}

// LoginRequest is used to authenticate dashboard users.
type RemoteToolData struct {
	ToolName string `json:"tool_name" binding:"required"`
	RemoteID string `json:"remote_id"`
	Version  string `json:"version"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest is used to create a new dashboard user.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}
