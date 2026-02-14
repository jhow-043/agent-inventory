// Package collector gathers system inventory data from Windows via WMI and the registry.
package collector

import (
	"fmt"
	"log/slog"

	"inventario/shared/dto"
)

// AgentVersion is the current version of the inventory agent.
const AgentVersion = "0.1.0"

// Collector orchestrates all inventory data collection.
type Collector struct {
	logger *slog.Logger
}

// New creates a new Collector with the given logger.
func New(logger *slog.Logger) *Collector {
	return &Collector{logger: logger}
}

// Collect gathers all inventory data and returns a complete InventoryRequest.
func (c *Collector) Collect() (*dto.InventoryRequest, error) {
	c.logger.Info("starting inventory collection")

	sys, err := c.collectSystem()
	if err != nil {
		return nil, fmt.Errorf("collect system info: %w", err)
	}

	hw, err := c.collectHardware()
	if err != nil {
		c.logger.Warn("failed to collect hardware info", "error", err)
		hw = &dto.HardwareData{}
	}

	disks, err := c.collectDisks()
	if err != nil {
		c.logger.Warn("failed to collect disk info", "error", err)
	}

	nics, err := c.collectNetwork()
	if err != nil {
		c.logger.Warn("failed to collect network info", "error", err)
	}

	software, err := c.collectSoftware()
	if err != nil {
		c.logger.Warn("failed to collect software info", "error", err)
	}

	license, err := c.collectLicense()
	if err != nil {
		c.logger.Warn("failed to collect license status", "error", err)
		license = "Unknown"
	}

	remoteTools := c.collectRemoteTools()

	req := &dto.InventoryRequest{
		Hostname:      sys.Hostname,
		SerialNumber:  sys.SerialNumber,
		OSName:        sys.OSName,
		OSVersion:     sys.OSVersion,
		OSBuild:       sys.OSBuild,
		OSArch:        sys.OSArch,
		LastBootTime:  sys.LastBootTime,
		LoggedInUser:  sys.LoggedInUser,
		AgentVersion:  AgentVersion,
		LicenseStatus: license,
		Hardware:      *hw,
		Disks:         disks,
		Network:       nics,
		Software:      software,
		RemoteTools:   remoteTools,
	}

	c.logger.Info("inventory collection complete",
		"hostname", req.Hostname,
		"serial_number", req.SerialNumber,
		"disks", len(disks),
		"network_interfaces", len(nics),
		"installed_software", len(software),
		"remote_tools", len(remoteTools),
	)

	return req, nil
}
