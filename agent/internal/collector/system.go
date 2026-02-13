package collector

import (
	"fmt"
	"os"
	"time"

	"github.com/yusufpapurcu/wmi"
)

// systemInfo holds the collected operating system and machine information.
type systemInfo struct {
	Hostname     string
	SerialNumber string
	OSName       string
	OSVersion    string
	OSBuild      string
	OSArch       string
	LastBootTime *time.Time
	LoggedInUser string
}

// win32OS maps fields from Win32_OperatingSystem.
type win32OS struct {
	Caption        string
	Version        string
	BuildNumber    string
	OSArchitecture string
	LastBootUpTime time.Time
}

// win32BIOS maps fields from Win32_BIOS (serial number).
type win32BIOS struct {
	SerialNumber string
}

// win32CS maps fields from Win32_ComputerSystem (logged-in user).
type win32CS struct {
	UserName string
}

// collectSystem gathers hostname, OS details, serial number, and logged-in user.
func (c *Collector) collectSystem() (*systemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	var osResult []win32OS
	if err := wmi.Query("SELECT Caption, Version, BuildNumber, OSArchitecture, LastBootUpTime FROM Win32_OperatingSystem", &osResult); err != nil {
		return nil, fmt.Errorf("query Win32_OperatingSystem: %w", err)
	}
	if len(osResult) == 0 {
		return nil, fmt.Errorf("no Win32_OperatingSystem results")
	}

	var bios []win32BIOS
	if err := wmi.Query("SELECT SerialNumber FROM Win32_BIOS", &bios); err != nil {
		return nil, fmt.Errorf("query Win32_BIOS: %w", err)
	}
	serial := ""
	if len(bios) > 0 {
		serial = bios[0].SerialNumber
	}

	var cs []win32CS
	if err := wmi.Query("SELECT UserName FROM Win32_ComputerSystem", &cs); err != nil {
		c.logger.Warn("failed to query logged-in user", "error", err)
	}
	user := ""
	if len(cs) > 0 {
		user = cs[0].UserName
	}

	bootTime := osResult[0].LastBootUpTime

	return &systemInfo{
		Hostname:     hostname,
		SerialNumber: serial,
		OSName:       osResult[0].Caption,
		OSVersion:    osResult[0].Version,
		OSBuild:      osResult[0].BuildNumber,
		OSArch:       osResult[0].OSArchitecture,
		LastBootTime: &bootTime,
		LoggedInUser: user,
	}, nil
}
