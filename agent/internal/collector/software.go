package collector

import (
	"strings"

	"golang.org/x/sys/windows/registry"

	"inventario/shared/dto"
)

// uninstallPaths are the registry paths that contain installed software entries.
var uninstallPaths = []string{
	`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
}

// collectSoftware reads installed software from the Windows registry.
func (c *Collector) collectSoftware() ([]dto.SoftwareData, error) {
	seen := make(map[string]bool)
	var software []dto.SoftwareData

	roots := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}

	for _, root := range roots {
		for _, path := range uninstallPaths {
			items, err := readUninstallKey(root, path)
			if err != nil {
				c.logger.Debug("failed to read registry key", "root", root, "path", path, "error", err)
				continue
			}
			for _, item := range items {
				key := strings.ToLower(item.Name + "|" + item.Version)
				if seen[key] {
					continue
				}
				seen[key] = true
				software = append(software, item)
			}
		}
	}

	return software, nil
}

// readUninstallKey enumerates software entries under the given registry key.
func readUninstallKey(root registry.Key, path string) ([]dto.SoftwareData, error) {
	key, err := registry.OpenKey(root, path, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	var software []dto.SoftwareData
	for _, name := range subkeys {
		sk, err := registry.OpenKey(key, name, registry.READ)
		if err != nil {
			continue
		}

		displayName, _, _ := sk.GetStringValue("DisplayName")
		if displayName == "" {
			sk.Close()
			continue
		}

		// Skip system components and updates.
		systemComponent, _, _ := sk.GetIntegerValue("SystemComponent")
		if systemComponent == 1 {
			sk.Close()
			continue
		}

		displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
		publisher, _, _ := sk.GetStringValue("Publisher")
		installDate, _, _ := sk.GetStringValue("InstallDate")

		software = append(software, dto.SoftwareData{
			Name:        displayName,
			Version:     displayVersion,
			Vendor:      publisher,
			InstallDate: installDate,
		})

		sk.Close()
	}

	return software, nil
}
