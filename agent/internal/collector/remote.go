package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"

	"inventario/shared/dto"
)

// collectRemoteTools detects installed remote access tools (TeamViewer, AnyDesk, RustDesk)
// and reads their remote IDs from the registry or configuration files.
func (c *Collector) collectRemoteTools() []dto.RemoteToolData {
	var tools []dto.RemoteToolData

	if t := c.detectTeamViewer(); t != nil {
		tools = append(tools, *t)
	}
	if t := c.detectAnyDesk(); t != nil {
		tools = append(tools, *t)
	}
	if t := c.detectRustDesk(); t != nil {
		tools = append(tools, *t)
	}

	return tools
}

// detectTeamViewer checks the registry for TeamViewer installation and reads the ClientID.
func (c *Collector) detectTeamViewer() *dto.RemoteToolData {
	paths := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\TeamViewer`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\TeamViewer`},
	}

	for _, p := range paths {
		key, err := registry.OpenKey(p.root, p.path, registry.READ)
		if err != nil {
			continue
		}
		defer key.Close()

		clientID, _, err := key.GetIntegerValue("ClientID")
		if err != nil {
			c.logger.Debug("TeamViewer registry found but no ClientID", "path", p.path, "error", err)
			continue
		}

		version := ""
		if v, _, err := key.GetStringValue("Version"); err == nil {
			version = v
		}

		c.logger.Info("detected TeamViewer", "client_id", clientID, "version", version)
		return &dto.RemoteToolData{
			ToolName: "TeamViewer",
			RemoteID: fmt.Sprintf("%d", clientID),
			Version:  version,
		}
	}

	// Also try to find version from uninstall registry if TeamViewer key was found
	return nil
}

// detectAnyDesk reads the AnyDesk ID from the system.conf file in ProgramData or user AppData.
func (c *Collector) detectAnyDesk() *dto.RemoteToolData {
	// First check if AnyDesk is installed by looking at uninstall keys
	version := c.getUninstallVersion("AnyDesk")
	if version == "" {
		// Try checking common install paths
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		for _, dir := range []string{programFiles, programFilesX86} {
			if dir == "" {
				continue
			}
			exePath := filepath.Join(dir, "AnyDesk", "AnyDesk.exe")
			if _, err := os.Stat(exePath); err == nil {
				version = "installed"
				break
			}
		}
	}

	// Look for AnyDesk ID in configuration files
	confPaths := []string{
		filepath.Join(os.Getenv("ProgramData"), "AnyDesk", "system.conf"),
		filepath.Join(os.Getenv("APPDATA"), "AnyDesk", "system.conf"),
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "AnyDesk", "system.conf"),
	}

	for _, confPath := range confPaths {
		if confPath == "" {
			continue
		}
		id := readAnyDeskID(confPath)
		if id != "" {
			c.logger.Info("detected AnyDesk", "remote_id", id, "version", version, "conf", confPath)
			return &dto.RemoteToolData{
				ToolName: "AnyDesk",
				RemoteID: id,
				Version:  version,
			}
		}
	}

	// AnyDesk installed but couldn't find ID
	if version != "" {
		c.logger.Info("detected AnyDesk but could not find remote ID", "version", version)
		return &dto.RemoteToolData{
			ToolName: "AnyDesk",
			RemoteID: "",
			Version:  version,
		}
	}

	return nil
}

// readAnyDeskID reads the ad.anynet.id value from an AnyDesk system.conf file.
func readAnyDeskID(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ad.anynet.id=") {
			return strings.TrimPrefix(line, "ad.anynet.id=")
		}
	}
	return ""
}

// detectRustDesk checks for RustDesk installation and reads its ID from config.
func (c *Collector) detectRustDesk() *dto.RemoteToolData {
	version := c.getUninstallVersion("RustDesk")

	// Check RustDesk config locations
	confPaths := []string{
		filepath.Join(os.Getenv("APPDATA"), "RustDesk", "config", "RustDesk.toml"),
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "RustDesk", "config", "RustDesk.toml"),
		filepath.Join(os.Getenv("ProgramData"), "RustDesk", "config", "RustDesk.toml"),
	}

	// Also try registry
	regPaths := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\RustDesk`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\RustDesk`},
		{registry.CURRENT_USER, `SOFTWARE\RustDesk`},
	}

	for _, p := range regPaths {
		key, err := registry.OpenKey(p.root, p.path, registry.READ)
		if err != nil {
			continue
		}
		if id, _, err := key.GetStringValue("ID"); err == nil && id != "" {
			if v, _, err := key.GetStringValue("Version"); err == nil {
				version = v
			}
			key.Close()
			c.logger.Info("detected RustDesk from registry", "remote_id", id, "version", version)
			return &dto.RemoteToolData{
				ToolName: "RustDesk",
				RemoteID: id,
				Version:  version,
			}
		}
		key.Close()
	}

	for _, confPath := range confPaths {
		if confPath == "" {
			continue
		}
		id := readRustDeskID(confPath)
		if id != "" {
			c.logger.Info("detected RustDesk", "remote_id", id, "version", version, "conf", confPath)
			return &dto.RemoteToolData{
				ToolName: "RustDesk",
				RemoteID: id,
				Version:  version,
			}
		}
	}

	// Fallback: run rustdesk.exe --get-id (works on RustDesk v1.4+ where config uses enc_id)
	if id := c.getRustDeskIDFromCLI(); id != "" {
		c.logger.Info("detected RustDesk via CLI --get-id", "remote_id", id, "version", version)
		return &dto.RemoteToolData{
			ToolName: "RustDesk",
			RemoteID: id,
			Version:  version,
		}
	}

	// Installed but no ID found
	if version != "" {
		c.logger.Info("detected RustDesk but could not find remote ID", "version", version)
		return &dto.RemoteToolData{
			ToolName: "RustDesk",
			RemoteID: "",
			Version:  version,
		}
	}

	return nil
}

// getRustDeskIDFromCLI runs "rustdesk.exe --get-id" and returns the ID from stdout.
// This is required for RustDesk v1.4+ where the config uses enc_id (encrypted) instead of plain id.
func (c *Collector) getRustDeskIDFromCLI() string {
	// Try common install locations
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "RustDesk", "rustdesk.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "RustDesk", "rustdesk.exe"),
	}

	// Also check uninstall registry for InstallLocation
	uninstallPaths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}
	for _, root := range []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER} {
		for _, path := range uninstallPaths {
			key, err := registry.OpenKey(root, path, registry.READ)
			if err != nil {
				continue
			}
			subkeys, _ := key.ReadSubKeyNames(-1)
			key.Close()
			for _, name := range subkeys {
				sk, err := registry.OpenKey(root, path+`\`+name, registry.READ)
				if err != nil {
					continue
				}
				displayName, _, _ := sk.GetStringValue("DisplayName")
				if strings.Contains(strings.ToLower(displayName), "rustdesk") {
					if loc, _, err := sk.GetStringValue("InstallLocation"); err == nil && loc != "" {
						candidates = append(candidates, filepath.Join(loc, "rustdesk.exe"))
					}
				}
				sk.Close()
			}
		}
	}

	for _, exePath := range candidates {
		if _, err := os.Stat(exePath); err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		out, err := exec.CommandContext(ctx, exePath, "--get-id").Output()
		cancel()
		if err != nil {
			c.logger.Debug("rustdesk --get-id failed", "path", exePath, "error", err)
			continue
		}
		id := strings.TrimSpace(string(out))
		if id != "" {
			return id
		}
	}
	return ""
}

// readRustDeskID reads the id field from a RustDesk TOML config file.
func readRustDeskID(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// TOML format: id = 'value' or id = "value" or id = value
		if strings.HasPrefix(line, "id") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			if key != "id" {
				continue
			}
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "'\"")
			if val != "" {
				return val
			}
		}
	}
	return ""
}

// getUninstallVersion looks up DisplayVersion in the Windows uninstall registry for a given software name.
func (c *Collector) getUninstallVersion(softwareName string) string {
	uninstallPaths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}
	roots := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}

	for _, root := range roots {
		for _, path := range uninstallPaths {
			key, err := registry.OpenKey(root, path, registry.READ)
			if err != nil {
				continue
			}
			subkeys, err := key.ReadSubKeyNames(-1)
			key.Close()
			if err != nil {
				continue
			}
			for _, name := range subkeys {
				sk, err := registry.OpenKey(root, path+`\`+name, registry.READ)
				if err != nil {
					continue
				}
				displayName, _, _ := sk.GetStringValue("DisplayName")
				if strings.Contains(strings.ToLower(displayName), strings.ToLower(softwareName)) {
					version, _, _ := sk.GetStringValue("DisplayVersion")
					sk.Close()
					if version != "" {
						return version
					}
				}
				sk.Close()
			}
		}
	}
	return ""
}
