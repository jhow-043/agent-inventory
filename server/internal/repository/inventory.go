package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/dto"
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

	// Upsert hardware
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

	// Replace disks
	if _, err = tx.ExecContext(ctx, "DELETE FROM disks WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete disks: %w", err)
	}
	for _, d := range req.Disks {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO disks (id, device_id, model, size_bytes, media_type, serial_number, interface_type)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6)
		`, deviceID, d.Model, d.SizeBytes, d.MediaType, d.SerialNumber, d.InterfaceType); err != nil {
			return fmt.Errorf("insert disk: %w", err)
		}
	}

	// Replace network interfaces
	if _, err = tx.ExecContext(ctx, "DELETE FROM network_interfaces WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete network interfaces: %w", err)
	}
	for _, n := range req.Network {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO network_interfaces (id, device_id, name, mac_address, ipv4_address, ipv6_address, speed_mbps, is_physical)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7)
		`, deviceID, n.Name, n.MACAddress, n.IPv4Address, n.IPv6Address, n.SpeedMbps, n.IsPhysical); err != nil {
			return fmt.Errorf("insert network interface: %w", err)
		}
	}

	// Replace installed software
	if _, err = tx.ExecContext(ctx, "DELETE FROM installed_software WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete installed software: %w", err)
	}
	for _, s := range req.Software {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO installed_software (id, device_id, name, version, vendor, install_date)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
		`, deviceID, s.Name, s.Version, s.Vendor, s.InstallDate); err != nil {
			return fmt.Errorf("insert software: %w", err)
		}
	}

	// Replace remote tools
	if _, err = tx.ExecContext(ctx, "DELETE FROM remote_tools WHERE device_id = $1", deviceID); err != nil {
		return fmt.Errorf("delete remote tools: %w", err)
	}
	for _, rt := range req.RemoteTools {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO remote_tools (id, device_id, tool_name, remote_id, version)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4)
		`, deviceID, rt.ToolName, rt.RemoteID, rt.Version); err != nil {
			return fmt.Errorf("insert remote tool: %w", err)
		}
	}

	return tx.Commit()
}
