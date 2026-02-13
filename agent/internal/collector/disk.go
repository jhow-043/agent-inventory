package collector

import (
	"fmt"

	"github.com/yusufpapurcu/wmi"
	"inventario/shared/dto"
)

// win32DiskDrive maps fields from Win32_DiskDrive.
type win32DiskDrive struct {
	Model         string
	Size          uint64
	MediaType     string
	SerialNumber  string
	InterfaceType string
}

// collectDisks gathers physical disk drive information via WMI.
func (c *Collector) collectDisks() ([]dto.DiskData, error) {
	var drives []win32DiskDrive
	if err := wmi.Query("SELECT Model, Size, MediaType, SerialNumber, InterfaceType FROM Win32_DiskDrive", &drives); err != nil {
		return nil, fmt.Errorf("query Win32_DiskDrive: %w", err)
	}

	disks := make([]dto.DiskData, 0, len(drives))
	for _, d := range drives {
		disks = append(disks, dto.DiskData{
			Model:         d.Model,
			SizeBytes:     int64(d.Size),
			MediaType:     classifyMediaType(d.MediaType),
			SerialNumber:  d.SerialNumber,
			InterfaceType: d.InterfaceType,
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
