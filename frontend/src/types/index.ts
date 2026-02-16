// API response and model types matching the Go backend DTOs.

export interface Device {
  id: string;
  hostname: string;
  serial_number: string;
  os_name: string;
  os_version: string;
  os_build: string;
  os_arch: string;
  last_boot_time: string | null;
  logged_in_user: string;
  agent_version: string;
  license_status: string;
  status: string;
  department_id: string | null;
  department_name: string | null;
  last_seen: string;
  created_at: string;
  updated_at: string;
}

export interface Hardware {
  id: string;
  device_id: string;
  cpu_model: string;
  cpu_cores: number;
  cpu_threads: number;
  ram_total_bytes: number;
  motherboard_manufacturer: string;
  motherboard_product: string;
  motherboard_serial: string;
  bios_vendor: string;
  bios_version: string;
}

export interface Disk {
  id: string;
  device_id: string;
  model: string;
  size_bytes: number;
  media_type: string;
  serial_number: string;
  interface_type: string;
  drive_letter: string;
  partition_size_bytes: number;
  free_space_bytes: number;
}

export interface NetworkInterface {
  id: string;
  device_id: string;
  name: string;
  mac_address: string;
  ipv4_address: string;
  ipv6_address: string;
  speed_mbps: number | null;
  is_physical: boolean;
}

export interface InstalledSoftware {
  id: string;
  device_id: string;
  name: string;
  version: string;
  vendor: string;
  install_date: string;
}

export interface RemoteTool {
  id: string;
  device_id: string;
  tool_name: string;
  remote_id: string;
  version: string;
}

export interface HardwareHistory {
  id: string;
  device_id: string;
  snapshot: string;
  changed_at: string;
}

export interface Department {
  id: string;
  name: string;
  created_at: string;
}

export interface DeviceListResponse {
  devices: Device[];
  total: number;
  page: number;
  limit: number;
}

export interface DashboardStats {
  total: number;
  online: number;
  offline: number;
  inactive: number;
}

export interface DeviceDetailResponse {
  device: Device;
  hardware: Hardware | null;
  disks: Disk[];
  network_interfaces: NetworkInterface[];
  installed_software: InstalledSoftware[];
  remote_tools: RemoteTool[];
  hardware_history: HardwareHistory[];
}

export interface DepartmentListResponse {
  departments: Department[];
  total: number;
}

export interface ErrorResponse {
  error: string;
}

export interface User {
  id: string;
  username: string;
  created_at: string;
}

export interface UserListResponse {
  users: User[];
  total: number;
}
