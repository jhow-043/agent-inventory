CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Dashboard users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Managed devices (Windows workstations)
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hostname VARCHAR(255) NOT NULL,
    serial_number VARCHAR(255) NOT NULL UNIQUE,
    os_name VARCHAR(100) NOT NULL DEFAULT '',
    os_version VARCHAR(100) NOT NULL DEFAULT '',
    os_build VARCHAR(50) NOT NULL DEFAULT '',
    os_arch VARCHAR(20) NOT NULL DEFAULT '',
    last_boot_time TIMESTAMPTZ,
    logged_in_user VARCHAR(255) NOT NULL DEFAULT '',
    agent_version VARCHAR(50) NOT NULL DEFAULT '',
    license_status VARCHAR(100) NOT NULL DEFAULT '',
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Agent authentication tokens (SHA-256 hashed)
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- CPU, RAM, motherboard and BIOS info
CREATE TABLE hardware (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    cpu_model VARCHAR(255) NOT NULL DEFAULT '',
    cpu_cores INTEGER NOT NULL DEFAULT 0,
    cpu_threads INTEGER NOT NULL DEFAULT 0,
    ram_total_bytes BIGINT NOT NULL DEFAULT 0,
    motherboard_manufacturer VARCHAR(255) NOT NULL DEFAULT '',
    motherboard_product VARCHAR(255) NOT NULL DEFAULT '',
    motherboard_serial VARCHAR(255) NOT NULL DEFAULT '',
    bios_vendor VARCHAR(255) NOT NULL DEFAULT '',
    bios_version VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Physical and logical disk drives
CREATE TABLE disks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    model VARCHAR(255) NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    media_type VARCHAR(20) NOT NULL DEFAULT '',
    serial_number VARCHAR(255) NOT NULL DEFAULT '',
    interface_type VARCHAR(20) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Network adapters
CREATE TABLE network_interfaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL DEFAULT '',
    mac_address VARCHAR(17) NOT NULL DEFAULT '',
    ipv4_address VARCHAR(15) NOT NULL DEFAULT '',
    ipv6_address VARCHAR(45) NOT NULL DEFAULT '',
    speed_mbps BIGINT,
    is_physical BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Installed applications
CREATE TABLE installed_software (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name VARCHAR(500) NOT NULL,
    version VARCHAR(100) NOT NULL DEFAULT '',
    vendor VARCHAR(255) NOT NULL DEFAULT '',
    install_date VARCHAR(20) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_devices_hostname ON devices(hostname);
CREATE INDEX idx_devices_last_seen ON devices(last_seen);
CREATE INDEX idx_disks_device_id ON disks(device_id);
CREATE INDEX idx_network_interfaces_device_id ON network_interfaces(device_id);
CREATE INDEX idx_installed_software_device_id ON installed_software(device_id);
CREATE INDEX idx_installed_software_device_name ON installed_software(device_id, name);
