package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/dto"
	"inventario/shared/models"
)

// hwChange represents a single granular hardware field change.
type hwChange struct {
	Component  string
	Field      string
	ChangeType string
	OldValue   string
	NewValue   string
}

// ──────────────────────────────────────────────────────────────────────────────
// CPU / RAM / Motherboard / BIOS — field-level comparison
// ──────────────────────────────────────────────────────────────────────────────

// detectHWFieldChanges compares the existing hardware record with incoming data
// and returns a slice of individual field changes.
func detectHWFieldChanges(existing models.Hardware, incoming dto.HardwareData) []hwChange {
	var changes []hwChange

	// CPU
	if existing.CPUModel != incoming.CPUModel {
		changes = append(changes, hwChange{"cpu", "model", "changed", existing.CPUModel, incoming.CPUModel})
	}
	if existing.CPUCores != incoming.CPUCores {
		changes = append(changes, hwChange{"cpu", "cores", "changed",
			fmt.Sprintf("%d", existing.CPUCores), fmt.Sprintf("%d", incoming.CPUCores)})
	}
	if existing.CPUThreads != incoming.CPUThreads {
		changes = append(changes, hwChange{"cpu", "threads", "changed",
			fmt.Sprintf("%d", existing.CPUThreads), fmt.Sprintf("%d", incoming.CPUThreads)})
	}

	// RAM
	if existing.RAMTotalBytes != incoming.RAMTotalBytes {
		changes = append(changes, hwChange{"ram", "total_bytes", "changed",
			formatBytesGo(existing.RAMTotalBytes), formatBytesGo(incoming.RAMTotalBytes)})
	}

	// Motherboard
	if existing.MotherboardManufacturer != incoming.MotherboardManufacturer {
		changes = append(changes, hwChange{"motherboard", "manufacturer", "changed",
			existing.MotherboardManufacturer, incoming.MotherboardManufacturer})
	}
	if existing.MotherboardProduct != incoming.MotherboardProduct {
		changes = append(changes, hwChange{"motherboard", "product", "changed",
			existing.MotherboardProduct, incoming.MotherboardProduct})
	}
	if existing.MotherboardSerial != incoming.MotherboardSerial {
		changes = append(changes, hwChange{"motherboard", "serial", "changed",
			existing.MotherboardSerial, incoming.MotherboardSerial})
	}

	// BIOS
	if existing.BIOSVendor != incoming.BIOSVendor {
		changes = append(changes, hwChange{"bios", "vendor", "changed",
			existing.BIOSVendor, incoming.BIOSVendor})
	}
	if existing.BIOSVersion != incoming.BIOSVersion {
		changes = append(changes, hwChange{"bios", "version", "changed",
			existing.BIOSVersion, incoming.BIOSVersion})
	}

	return changes
}

// saveHWHistory persists a list of hardware field changes to hardware_history.
func saveHWHistory(ctx context.Context, tx *sqlx.Tx, deviceID uuid.UUID, existing models.Hardware, changes []hwChange) error {
	snapshot, _ := json.Marshal(existing)
	snapshotStr := string(snapshot)

	for _, ch := range changes {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
			 VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, NOW())`,
			deviceID, snapshotStr, ch.Component, ch.ChangeType, ch.Field, ch.OldValue, ch.NewValue); err != nil {
			return fmt.Errorf("save hardware history (%s.%s): %w", ch.Component, ch.Field, err)
		}
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Disk changes — compare by serial (fallback to composite key)
// ──────────────────────────────────────────────────────────────────────────────

// diskKey builds a unique lookup key for a disk.
// Uses serial_number when available; otherwise falls back to a composite of
// model + size + media_type to avoid collisions when multiple disks have empty
// serials (e.g. USB drives, virtual disks).
func diskKey(serial, model, mediaType string, sizeBytes int64) string {
	s := strings.TrimSpace(serial)
	if s != "" {
		return s
	}
	return fmt.Sprintf("model:%s|size:%d|type:%s",
		strings.TrimSpace(model), sizeBytes, strings.TrimSpace(mediaType))
}

// saveDiskHistory detects and persists disk-level changes (added/removed/changed).
func saveDiskHistory(ctx context.Context, tx *sqlx.Tx, deviceID uuid.UUID,
	currentDisks []models.Disk, incomingDisks []dto.DiskData) error {

	currentMap := make(map[string]models.Disk, len(currentDisks))
	for _, d := range currentDisks {
		currentMap[diskKey(d.SerialNumber, d.Model, d.MediaType, d.SizeBytes)] = d
	}
	incomingMap := make(map[string]dto.DiskData, len(incomingDisks))
	for _, d := range incomingDisks {
		incomingMap[diskKey(d.SerialNumber, d.Model, d.MediaType, d.SizeBytes)] = d
	}

	snapshot, _ := json.Marshal(currentDisks)
	snapshotStr := string(snapshot)

	// Detect new disks
	for key, d := range incomingMap {
		if _, exists := currentMap[key]; !exists {
			desc := fmt.Sprintf("%s (%s)", strings.TrimSpace(d.Model), formatBytesGo(d.SizeBytes))
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
				 VALUES (uuid_generate_v4(), $1, $2, 'disk', 'added', 'disk', '', $3, NOW())`,
				deviceID, snapshotStr, desc); err != nil {
				return fmt.Errorf("save disk added history: %w", err)
			}
		}
	}

	// Detect removed disks
	for key, d := range currentMap {
		if _, exists := incomingMap[key]; !exists {
			desc := fmt.Sprintf("%s (%s)", strings.TrimSpace(d.Model), formatBytesGo(d.SizeBytes))
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
				 VALUES (uuid_generate_v4(), $1, $2, 'disk', 'removed', 'disk', $3, '', NOW())`,
				deviceID, snapshotStr, desc); err != nil {
				return fmt.Errorf("save disk removed history: %w", err)
			}
		}
	}

	// Detect changed disks (same key, different size or media type)
	for key, curr := range currentMap {
		if inc, exists := incomingMap[key]; exists {
			if curr.SizeBytes != inc.SizeBytes {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
					 VALUES (uuid_generate_v4(), $1, $2, 'disk', 'changed', 'size_bytes', $3, $4, NOW())`,
					deviceID, snapshotStr,
					fmt.Sprintf("%s: %s", strings.TrimSpace(curr.Model), formatBytesGo(curr.SizeBytes)),
					fmt.Sprintf("%s: %s", strings.TrimSpace(curr.Model), formatBytesGo(inc.SizeBytes))); err != nil {
					return fmt.Errorf("save disk change history: %w", err)
				}
			}
			if curr.MediaType != inc.MediaType && inc.MediaType != "" {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
					 VALUES (uuid_generate_v4(), $1, $2, 'disk', 'changed', 'media_type', $3, $4, NOW())`,
					deviceID, snapshotStr,
					fmt.Sprintf("%s: %s", strings.TrimSpace(curr.Model), curr.MediaType),
					fmt.Sprintf("%s: %s", strings.TrimSpace(curr.Model), inc.MediaType)); err != nil {
					return fmt.Errorf("save disk media type change: %w", err)
				}
			}
		}
	}

	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Network interface changes — compare by MAC address
// ──────────────────────────────────────────────────────────────────────────────

// nicKey builds a unique lookup key for a network interface.
func nicKey(mac, name string) string {
	key := strings.TrimSpace(strings.ToLower(mac))
	if key == "" {
		key = "name:" + strings.TrimSpace(name)
	}
	return key
}

// saveNICHistory detects and persists network interface changes (added/removed).
func saveNICHistory(ctx context.Context, tx *sqlx.Tx, deviceID uuid.UUID,
	currentNICs []models.NetworkInterface, incomingNICs []dto.NetworkData) error {

	currentMap := make(map[string]models.NetworkInterface, len(currentNICs))
	for _, n := range currentNICs {
		currentMap[nicKey(n.MACAddress, n.Name)] = n
	}
	incomingMap := make(map[string]dto.NetworkData, len(incomingNICs))
	for _, n := range incomingNICs {
		incomingMap[nicKey(n.MACAddress, n.Name)] = n
	}

	snapshot, _ := json.Marshal(currentNICs)
	snapshotStr := string(snapshot)

	// Detect new interfaces
	for key, n := range incomingMap {
		if _, exists := currentMap[key]; !exists {
			desc := strings.TrimSpace(n.Name)
			if n.MACAddress != "" {
				desc += " (" + n.MACAddress + ")"
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
				 VALUES (uuid_generate_v4(), $1, $2, 'network', 'added', 'interface', '', $3, NOW())`,
				deviceID, snapshotStr, desc); err != nil {
				return fmt.Errorf("save NIC added history: %w", err)
			}
		}
	}

	// Detect removed interfaces
	for key, n := range currentMap {
		if _, exists := incomingMap[key]; !exists {
			desc := strings.TrimSpace(n.Name)
			if n.MACAddress != "" {
				desc += " (" + n.MACAddress + ")"
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO hardware_history (id, device_id, snapshot, component, change_type, field, old_value, new_value, changed_at)
				 VALUES (uuid_generate_v4(), $1, $2, 'network', 'removed', 'interface', $3, '', NOW())`,
				deviceID, snapshotStr, desc); err != nil {
				return fmt.Errorf("save NIC removed history: %w", err)
			}
		}
	}

	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// formatBytesGo converts bytes to a human-readable string (e.g. "16 GB").
// Returns "N/A" for values ≤ 0.
func formatBytesGo(b int64) string {
	if b <= 0 {
		return "N/A"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	sizes := []string{"KB", "MB", "GB", "TB"}
	val := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %s", val, sizes[exp])
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}
