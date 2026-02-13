package collector

import (
	"fmt"
	"net"
	"strings"

	"github.com/yusufpapurcu/wmi"
	"inventario/shared/dto"
)

// win32NetworkAdapter maps fields from Win32_NetworkAdapter.
type win32NetworkAdapter struct {
	Name            string
	MACAddress      string
	Speed           *uint64
	PhysicalAdapter bool
}

// collectNetwork gathers physical network adapter info from WMI and IP addresses from Go's net package.
func (c *Collector) collectNetwork() ([]dto.NetworkData, error) {
	var adapters []win32NetworkAdapter
	if err := wmi.Query(
		"SELECT Name, MACAddress, Speed, PhysicalAdapter FROM Win32_NetworkAdapter WHERE PhysicalAdapter = TRUE AND MACAddress IS NOT NULL",
		&adapters,
	); err != nil {
		return nil, fmt.Errorf("query Win32_NetworkAdapter: %w", err)
	}

	// Build a map of MAC → (ipv4, ipv6) from Go's net package for reliable IP resolution.
	macIPMap := buildMACIPMap()

	nics := make([]dto.NetworkData, 0, len(adapters))
	for _, a := range adapters {
		mac := normalizeMAC(a.MACAddress)

		ipv4, ipv6 := "", ""
		if ips, ok := macIPMap[mac]; ok {
			ipv4 = ips[0]
			ipv6 = ips[1]
		}

		var speedMbps *int
		if a.Speed != nil && *a.Speed > 0 {
			s := int(*a.Speed / 1_000_000)
			speedMbps = &s
		}

		nics = append(nics, dto.NetworkData{
			Name:        a.Name,
			MACAddress:  a.MACAddress,
			IPv4Address: ipv4,
			IPv6Address: ipv6,
			SpeedMbps:   speedMbps,
			IsPhysical:  true,
		})
	}

	return nics, nil
}

// buildMACIPMap returns a map from uppercase MAC addresses to [ipv4, ipv6] pairs.
func buildMACIPMap() map[string][2]string {
	result := make(map[string][2]string)

	ifaces, err := net.Interfaces()
	if err != nil {
		return result
	}

	for _, iface := range ifaces {
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		mac := normalizeMAC(iface.HardwareAddr.String())

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipv4, ipv6 string
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() {
				continue
			}
			if ipNet.IP.To4() != nil && ipv4 == "" {
				ipv4 = ipNet.IP.String()
			} else if ipNet.IP.To4() == nil && ipv6 == "" {
				ipv6 = ipNet.IP.String()
			}
		}

		result[mac] = [2]string{ipv4, ipv6}
	}

	return result
}

// normalizeMAC uppercases the MAC address for consistent map lookups.
func normalizeMAC(mac string) string {
	return strings.ToUpper(mac)
}
