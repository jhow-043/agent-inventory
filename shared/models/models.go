// Package models defines the database entities for the inventory system.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a managed Windows workstation.
type Device struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Hostname       string     `json:"hostname" db:"hostname"`
	SerialNumber   string     `json:"serial_number" db:"serial_number"`
	OSName         string     `json:"os_name" db:"os_name"`
	OSVersion      string     `json:"os_version" db:"os_version"`
	OSBuild        string     `json:"os_build" db:"os_build"`
	OSArch         string     `json:"os_arch" db:"os_arch"`
	LastBootTime   *time.Time `json:"last_boot_time,omitempty" db:"last_boot_time"`
	LoggedInUser   string     `json:"logged_in_user" db:"logged_in_user"`
	AgentVersion   string     `json:"agent_version" db:"agent_version"`
	LicenseStatus  string     `json:"license_status" db:"license_status"`
	Status         string     `json:"status" db:"status"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	DepartmentName *string    `json:"department_name,omitempty" db:"department_name"`
	LastSeen       time.Time  `json:"last_seen" db:"last_seen"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// DeviceToken stores the hashed authentication token for an agent.
type DeviceToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DeviceID  uuid.UUID `json:"device_id" db:"device_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Hardware stores CPU, RAM, motherboard, and BIOS information.
type Hardware struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	DeviceID                uuid.UUID `json:"device_id" db:"device_id"`
	CPUModel                string    `json:"cpu_model" db:"cpu_model"`
	CPUCores                int       `json:"cpu_cores" db:"cpu_cores"`
	CPUThreads              int       `json:"cpu_threads" db:"cpu_threads"`
	RAMTotalBytes           int64     `json:"ram_total_bytes" db:"ram_total_bytes"`
	MotherboardManufacturer string    `json:"motherboard_manufacturer" db:"motherboard_manufacturer"`
	MotherboardProduct      string    `json:"motherboard_product" db:"motherboard_product"`
	MotherboardSerial       string    `json:"motherboard_serial" db:"motherboard_serial"`
	BIOSVendor              string    `json:"bios_vendor" db:"bios_vendor"`
	BIOSVersion             string    `json:"bios_version" db:"bios_version"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// Disk represents a physical or logical disk drive.
type Disk struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	DeviceID           uuid.UUID `json:"device_id" db:"device_id"`
	Model              string    `json:"model" db:"model"`
	SizeBytes          int64     `json:"size_bytes" db:"size_bytes"`
	MediaType          string    `json:"media_type" db:"media_type"`
	SerialNumber       string    `json:"serial_number" db:"serial_number"`
	InterfaceType      string    `json:"interface_type" db:"interface_type"`
	DriveLetter        string    `json:"drive_letter" db:"drive_letter"`
	PartitionSizeBytes int64     `json:"partition_size_bytes" db:"partition_size_bytes"`
	FreeSpaceBytes     int64     `json:"free_space_bytes" db:"free_space_bytes"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// NetworkInterface represents a network adapter.
type NetworkInterface struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DeviceID    uuid.UUID `json:"device_id" db:"device_id"`
	Name        string    `json:"name" db:"name"`
	MACAddress  string    `json:"mac_address" db:"mac_address"`
	IPv4Address string    `json:"ipv4_address" db:"ipv4_address"`
	IPv6Address string    `json:"ipv6_address" db:"ipv6_address"`
	SpeedMbps   *int      `json:"speed_mbps,omitempty" db:"speed_mbps"`
	IsPhysical  bool      `json:"is_physical" db:"is_physical"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// InstalledSoftware represents an installed application.
type InstalledSoftware struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DeviceID    uuid.UUID `json:"device_id" db:"device_id"`
	Name        string    `json:"name" db:"name"`
	Version     string    `json:"version" db:"version"`
	Vendor      string    `json:"vendor" db:"vendor"`
	InstallDate string    `json:"install_date" db:"install_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RemoteTool represents a remote access tool installed on a device.
type RemoteTool struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DeviceID  uuid.UUID `json:"device_id" db:"device_id"`
	ToolName  string    `json:"tool_name" db:"tool_name"`
	RemoteID  string    `json:"remote_id" db:"remote_id"`
	Version   string    `json:"version" db:"version"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Department represents an organizational unit that devices can be assigned to.
type Department struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// HardwareHistory stores a snapshot of hardware state before it changed,
// along with structured change details (component, field, old/new values).
type HardwareHistory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	DeviceID   uuid.UUID `json:"device_id" db:"device_id"`
	Snapshot   string    `json:"snapshot" db:"snapshot"`
	Component  *string   `json:"component,omitempty" db:"component"`     // cpu, ram, motherboard, bios, disk, network
	ChangeType *string   `json:"change_type,omitempty" db:"change_type"` // changed, added, removed
	Field      *string   `json:"field,omitempty" db:"field"`             // e.g. model, total_bytes
	OldValue   *string   `json:"old_value,omitempty" db:"old_value"`
	NewValue   *string   `json:"new_value,omitempty" db:"new_value"`
	ChangedAt  time.Time `json:"changed_at" db:"changed_at"`
}

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Name         string    `json:"name" db:"name"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog represents a record of an important system action.
type AuditLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       *uuid.UUID `json:"user_id" db:"user_id"`
	Username     string     `json:"username" db:"username"`
	Action       string     `json:"action" db:"action"`
	ResourceType string     `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty" db:"resource_id"`
	Details      string     `json:"details,omitempty" db:"details"` // JSONB stored as string
	IPAddress    string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// DeviceActivityLog tracks changes detected during inventory submissions.
type DeviceActivityLog struct {
	ID           uuid.UUID `json:"id" db:"id"`
	DeviceID     uuid.UUID `json:"device_id" db:"device_id"`
	ActivityType string    `json:"activity_type" db:"activity_type"`
	Description  string    `json:"description" db:"description"`
	OldValue     *string   `json:"old_value,omitempty" db:"old_value"`
	NewValue     *string   `json:"new_value,omitempty" db:"new_value"`
	Metadata     *string   `json:"metadata,omitempty" db:"metadata"` // JSONB
	DetectedAt   time.Time `json:"detected_at" db:"detected_at"`
}
