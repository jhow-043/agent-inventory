package collector

import (
	"fmt"

	"inventario/shared/dto"

	"github.com/yusufpapurcu/wmi"
)

// win32Processor maps fields from Win32_Processor.
type win32Processor struct {
	Name                      string
	NumberOfCores             uint32
	NumberOfLogicalProcessors uint32
}

// win32PhysicalMemory maps the Capacity field from Win32_PhysicalMemory.
type win32PhysicalMemory struct {
	Capacity uint64
}

// win32CSMemory maps TotalPhysicalMemory from Win32_ComputerSystem (fallback for VMs).
type win32CSMemory struct {
	TotalPhysicalMemory uint64
}

// win32BaseBoard maps fields from Win32_BaseBoard.
type win32BaseBoard struct {
	Manufacturer string
	Product      string
	SerialNumber string
}

// win32BIOSDetail maps fields from Win32_BIOS for vendor and version.
type win32BIOSDetail struct {
	Manufacturer      string
	SMBIOSBIOSVersion string
}

// collectHardware gathers CPU, RAM, motherboard, and BIOS information via WMI.
func (c *Collector) collectHardware() (*dto.HardwareData, error) {
	// CPU
	var procs []win32Processor
	if err := wmi.Query("SELECT Name, NumberOfCores, NumberOfLogicalProcessors FROM Win32_Processor", &procs); err != nil {
		return nil, fmt.Errorf("query Win32_Processor: %w", err)
	}
	cpuModel := ""
	var cpuCores, cpuThreads int
	if len(procs) > 0 {
		cpuModel = procs[0].Name
		cpuCores = int(procs[0].NumberOfCores)
		cpuThreads = int(procs[0].NumberOfLogicalProcessors)
	}

	// RAM — sum all physical memory sticks
	var mem []win32PhysicalMemory
	if err := wmi.Query("SELECT Capacity FROM Win32_PhysicalMemory", &mem); err != nil {
		c.logger.Warn("failed to query Win32_PhysicalMemory", "error", err)
	}
	var totalRAM int64
	for _, m := range mem {
		totalRAM += int64(m.Capacity)
	}
	// Fallback for VMs where Win32_PhysicalMemory may return empty or zero
	if totalRAM == 0 {
		var csm []win32CSMemory
		if err := wmi.Query("SELECT TotalPhysicalMemory FROM Win32_ComputerSystem", &csm); err != nil {
			c.logger.Warn("failed to query Win32_ComputerSystem for RAM fallback", "error", err)
		} else if len(csm) > 0 {
			totalRAM = int64(csm[0].TotalPhysicalMemory)
			c.logger.Info("used Win32_ComputerSystem fallback for RAM", "total_bytes", totalRAM)
		}
	}

	// Motherboard
	var boards []win32BaseBoard
	if err := wmi.Query("SELECT Manufacturer, Product, SerialNumber FROM Win32_BaseBoard", &boards); err != nil {
		c.logger.Warn("failed to query motherboard", "error", err)
	}
	mbMfg, mbProduct, mbSerial := "", "", ""
	if len(boards) > 0 {
		mbMfg = boards[0].Manufacturer
		mbProduct = boards[0].Product
		mbSerial = boards[0].SerialNumber
	}

	// BIOS
	var biosInfo []win32BIOSDetail
	if err := wmi.Query("SELECT Manufacturer, SMBIOSBIOSVersion FROM Win32_BIOS", &biosInfo); err != nil {
		c.logger.Warn("failed to query BIOS details", "error", err)
	}
	biosVendor, biosVersion := "", ""
	if len(biosInfo) > 0 {
		biosVendor = biosInfo[0].Manufacturer
		biosVersion = biosInfo[0].SMBIOSBIOSVersion
	}

	return &dto.HardwareData{
		CPUModel:                cpuModel,
		CPUCores:                cpuCores,
		CPUThreads:              cpuThreads,
		RAMTotalBytes:           totalRAM,
		MotherboardManufacturer: mbMfg,
		MotherboardProduct:      mbProduct,
		MotherboardSerial:       mbSerial,
		BIOSVendor:              biosVendor,
		BIOSVersion:             biosVersion,
	}, nil
}
