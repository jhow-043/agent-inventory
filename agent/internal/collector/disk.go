package collector

import (
	"fmt"

	"inventario/shared/dto"

	"github.com/yusufpapurcu/wmi"
)

// win32DiskDrive maps fields from Win32_DiskDrive.
type win32DiskDrive struct {
	Model         string
	Size          uint64
	MediaType     string
	SerialNumber  string
	InterfaceType string
}

// win32LogicalDisk maps fields from Win32_LogicalDisk (local fixed drives).
type win32LogicalDisk struct {
	DeviceID  string // Drive letter, e.g. "C:"
	Size      uint64
	FreeSpace uint64
}

// collectDisks gathers physical disk drive and logical partition info via WMI.
func (c *Collector) collectDisks() ([]dto.DiskData, error) {
	var drives []win32DiskDrive
	if err := wmi.Query("SELECT Model, Size, MediaType, SerialNumber, InterfaceType FROM Win32_DiskDrive", &drives); err != nil {
		return nil, fmt.Errorf("query Win32_DiskDrive: %w", err)
	}

	// Query logical disks (DriveType=3 means local fixed disk).
	var partitions []win32LogicalDisk
	if err := wmi.Query("SELECT DeviceID, Size, FreeSpace FROM Win32_LogicalDisk WHERE DriveType = 3", &partitions); err != nil {
		// Non-fatal: continue without partition info.
		partitions = nil
	}

	// Build map of drive letter → partition info.
	partMap := make(map[string]win32LogicalDisk, len(partitions))
	for _, p := range partitions {
		partMap[p.DeviceID] = p
	}

	disks := make([]dto.DiskData, 0, len(drives)+len(partitions))

	// Add physical drive entries.
	for _, d := range drives {
		disks = append(disks, dto.DiskData{
			Model:         d.Model,
			SizeBytes:     int64(d.Size),
			MediaType:     classifyMediaType(d.MediaType),
			SerialNumber:  d.SerialNumber,
			InterfaceType: d.InterfaceType,
		})
	}

	// Add logical partition entries (with free space info).
	for _, p := range partitions {
		disks = append(disks, dto.DiskData{
			Model:              fmt.Sprintf("Partition %s", p.DeviceID),
			SizeBytes:          int64(p.Size),
			MediaType:          "Partition",
			DriveLetter:        p.DeviceID,
			PartitionSizeBytes: int64(p.Size),
			FreeSpaceBytes:     int64(p.FreeSpace),
		})
	}

	return disks, nil
}

// classifyMediaType normalizes the WMI MediaType string into a simpler label.
func classifyMediaType(raw string) string {
	switch raw {
	case "Fixed hard disk media":
		return "HDD"
	case "Removable Media":
		return "Removable"
	default:
		// WMI often returns empty for NVMe/SSD drives
		if raw == "" {
			return "SSD"
		}
		return raw
	}
}
