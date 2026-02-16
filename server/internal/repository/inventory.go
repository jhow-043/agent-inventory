package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/dto"
	"inventario/shared/models"
)

// InventoryRepository handles the transactional upsert of a full inventory snapshot.
type InventoryRepository struct {
	db *sqlx.DB
}

// NewInventoryRepository creates a new InventoryRepository.
func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Save persists an entire inventory snapshot inside a single database transaction.
// It upserts the device and hardware rows, then replaces disks, NICs and software.
func (r *InventoryRepository) Save(ctx context.Context, deviceID uuid.UUID, req *dto.InventoryRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Upsert device
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO devices (id, hostname, serial_number, os_name, os_version, os_build, os_arch,
			last_boot_time, logged_in_user, agent_version, license_status, last_seen, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			hostname       = EXCLUDED.hostname,
			os_name        = EXCLUDED.os_name,
			os_version     = EXCLUDED.os_version,
			os_build       = EXCLUDED.os_build,
			os_arch        = EXCLUDED.os_arch,
			last_boot_time = EXCLUDED.last_boot_time,
			logged_in_user = EXCLUDED.logged_in_user,
			agent_version  = EXCLUDED.agent_version,
			license_status = EXCLUDED.license_status,
			last_seen      = NOW(),
			updated_at     = NOW()
	`, deviceID, req.Hostname, req.SerialNumber,
		req.OSName, req.OSVersion, req.OSBuild, req.OSArch,
		req.LastBootTime, req.LoggedInUser, req.AgentVersion, req.LicenseStatus,
	); err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}

	// Upsert hardware — save a snapshot first if hardware changed.
	var existingHW models.Hardware
	hwErr := tx.GetContext(ctx, &existingHW, "SELECT * FROM hardware WHERE device_id = $1", deviceID)
	if hwErr == nil {
		// Hardware row exists — compare key fields.
		incoming := req.Hardware
		changed := existingHW.CPUModel != incoming.CPUModel ||
			existingHW.CPUCores != incoming.CPUCores ||
			existingHW.CPUThreads != incoming.CPUThreads ||
			existingHW.RAMTotalBytes != incoming.RAMTotalBytes ||
			existingHW.MotherboardManufacturer != incoming.MotherboardManufacturer ||
			existingHW.MotherboardProduct != incoming.MotherboardProduct ||
			existingHW.MotherboardSerial != incoming.MotherboardSerial ||
			existingHW.BIOSVendor != incoming.BIOSVendor ||
			existingHW.BIOSVersion != incoming.BIOSVersion

		if changed {
			snapshot, _ := json.Marshal(existingHW)
			if _, err = tx.ExecContext(ctx,
				"INSERT INTO hardware_history (id, device_id, snapshot, changed_at) VALUES (uuid_generate_v4(), $1, $2, NOW())",
				deviceID, string(snapshot)); err != nil {
				return fmt.Errorf("save hardware history: %w", err)
			}
		}
	} else if hwErr != sql.ErrNoRows {
		return fmt.Errorf("check existing hardware: %w", hwErr)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO hardware (id, device_id, cpu_model, cpu_cores, cpu_threads, ram_total_bytes,
			motherboard_manufacturer, motherboard_product, motherboard_serial,
			bios_vendor, bios_version, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (device_id) DO UPDATE SET
			cpu_model                = EXCLUDED.cpu_model,
			cpu_cores                = EXCLUDED.cpu_cores,
			cpu_threads              = EXCLUDED.cpu_threads,
			ram_total_bytes          = EXCLUDED.ram_total_bytes,
			motherboard_manufacturer = EXCLUDED.motherboard_manufacturer,
			motherboard_product      = EXCLUDED.motherboard_product,
			motherboard_serial       = EXCLUDED.motherboard_serial,
			bios_vendor              = EXCLUDED.bios_vendor,
			bios_version             = EXCLUDED.bios_version,
			updated_at               = NOW()
	`, deviceID,
		req.Hardware.CPUModel, req.Hardware.CPUCores, req.Hardware.CPUThreads,
		req.Hardware.RAMTotalBytes,
		req.Hardware.MotherboardManufacturer, req.Hardware.MotherboardProduct,
		req.Hardware.MotherboardSerial,
		req.Hardware.BIOSVendor, req.Hardware.BIOSVersion,
	); err != nil {
		return fmt.Errorf("upsert hardware: %w", err)
	}

	// Replace disks (batch insert)
	if _, err = tx.ExecContext(ctx, "DELETE FROM disks WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete disks: %w", err)
	}
	if len(req.Disks) > 0 {
		vals := make([]string, 0, len(req.Disks))
		args := []interface{}{}
		for i, d := range req.Disks {
			base := i*9 + 1
			vals = append(vals, fmt.Sprintf("(uuid_generate_v4(), $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8))
			args = append(args, deviceID, d.Model, d.SizeBytes, d.MediaType, d.SerialNumber, d.InterfaceType,
				d.DriveLetter, d.PartitionSizeBytes, d.FreeSpaceBytes)
		}
		q := "INSERT INTO disks (id, device_id, model, size_bytes, media_type, serial_number, interface_type, drive_letter, partition_size_bytes, free_space_bytes) VALUES " + strings.Join(vals, ", ")
		if _, err = tx.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("batch insert disks: %w", err)
		}
	}

	// Replace network interfaces (batch insert)
	if _, err = tx.ExecContext(ctx, "DELETE FROM network_interfaces WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete network interfaces: %w", err)
	}
	if len(req.Network) > 0 {
		vals := make([]string, 0, len(req.Network))
		args := []interface{}{}
		for i, n := range req.Network {
			base := i*7 + 1
			vals = append(vals, fmt.Sprintf("(uuid_generate_v4(), $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				base, base+1, base+2, base+3, base+4, base+5, base+6))
			args = append(args, deviceID, n.Name, n.MACAddress, n.IPv4Address, n.IPv6Address, n.SpeedMbps, n.IsPhysical)
		}
		q := "INSERT INTO network_interfaces (id, device_id, name, mac_address, ipv4_address, ipv6_address, speed_mbps, is_physical) VALUES " + strings.Join(vals, ", ")
		if _, err = tx.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("batch insert network interfaces: %w", err)
		}
	}

	// Replace installed software (batch insert — chunked to avoid param limit)
	if _, err = tx.ExecContext(ctx, "DELETE FROM installed_software WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete installed software: %w", err)
	}
	if len(req.Software) > 0 {
		const chunkSize = 200 // ~1000 params per chunk (5 fields each)
		for start := 0; start < len(req.Software); start += chunkSize {
			end := start + chunkSize
			if end > len(req.Software) {
				end = len(req.Software)
			}
			chunk := req.Software[start:end]
			vals := make([]string, 0, len(chunk))
			args := []interface{}{}
			for i, s := range chunk {
				base := i*5 + 1
				vals = append(vals, fmt.Sprintf("(uuid_generate_v4(), $%d, $%d, $%d, $%d, $%d)",
					base, base+1, base+2, base+3, base+4))
				args = append(args, deviceID, s.Name, s.Version, s.Vendor, s.InstallDate)
			}
			q := "INSERT INTO installed_software (id, device_id, name, version, vendor, install_date) VALUES " + strings.Join(vals, ", ")
			if _, err = tx.ExecContext(ctx, q, args...); err != nil {
				return fmt.Errorf("batch insert software: %w", err)
			}
		}
	}

	// Replace remote tools (batch insert)
	if _, err = tx.ExecContext(ctx, "DELETE FROM remote_tools WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete remote tools: %w", err)
	}
	if len(req.RemoteTools) > 0 {
		vals := make([]string, 0, len(req.RemoteTools))
		args := []interface{}{}
		for i, rt := range req.RemoteTools {
			base := i*4 + 1
			vals = append(vals, fmt.Sprintf("(uuid_generate_v4(), $%d, $%d, $%d, $%d)",
				base, base+1, base+2, base+3))
			args = append(args, deviceID, rt.ToolName, rt.RemoteID, rt.Version)
		}
		q := "INSERT INTO remote_tools (id, device_id, tool_name, remote_id, version) VALUES " + strings.Join(vals, ", ")
		if _, err = tx.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("batch insert remote tools: %w", err)
		}
	}

	return tx.Commit()
}
