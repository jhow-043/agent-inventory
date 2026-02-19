// Device API calls.

import { request } from './client';
import type { DeviceListResponse, DeviceDetailResponse, DeviceActivityResponse, HardwareHistoryResponse } from '../types';

export interface DeviceListParams {
  hostname?: string;
  os?: string;
  status?: string;
  department_id?: string;
  sort?: string;
  order?: string;
  page?: number;
  limit?: number;
}

function buildDeviceQS(params: DeviceListParams): string {
  const qs = new URLSearchParams();
  if (params.hostname) qs.set('hostname', params.hostname);
  if (params.os) qs.set('os', params.os);
  if (params.status) qs.set('status', params.status);
  if (params.department_id) qs.set('department_id', params.department_id);
  if (params.sort) qs.set('sort', params.sort);
  if (params.order) qs.set('order', params.order);
  if (params.page) qs.set('page', String(params.page));
  if (params.limit) qs.set('limit', String(params.limit));
  return qs.toString();
}

export async function getDevices(params: DeviceListParams = {}): Promise<DeviceListResponse> {
  const q = buildDeviceQS(params);
  return request<DeviceListResponse>(`/devices${q ? `?${q}` : ''}`);
}

export async function getDevice(id: string): Promise<DeviceDetailResponse> {
  return request<DeviceDetailResponse>(`/devices/${id}`);
}

export async function updateDeviceStatus(id: string, status: 'active' | 'inactive'): Promise<void> {
  await request(`/devices/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  });
}

export async function updateDeviceDepartment(id: string, departmentId: string | null): Promise<void> {
  await request(`/devices/${id}/department`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ department_id: departmentId }),
  });
}

export async function getHardwareHistory(id: string, page = 1, limit = 50, component = ''): Promise<HardwareHistoryResponse> {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (component) params.set('component', component);
  return request<HardwareHistoryResponse>(`/devices/${id}/hardware-history?${params.toString()}`);
}

export async function getDeviceActivity(id: string, page = 1, limit = 50): Promise<DeviceActivityResponse> {
  return request<DeviceActivityResponse>(`/devices/${id}/activity?page=${page}&limit=${limit}`);
}

export async function deleteDevice(id: string): Promise<{ message: string }> {
  return request<{ message: string }>(`/devices/${id}`, {
    method: 'DELETE',
  });
}

// Bulk actions

export interface BulkActionResponse {
  affected: number;
  message: string;
}

export async function bulkUpdateStatus(deviceIds: string[], status: 'active' | 'inactive'): Promise<BulkActionResponse> {
  return request<BulkActionResponse>('/devices/bulk/status', {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_ids: deviceIds, status }),
  });
}

export async function bulkUpdateDepartment(deviceIds: string[], departmentId: string | null): Promise<BulkActionResponse> {
  return request<BulkActionResponse>('/devices/bulk/department', {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_ids: deviceIds, department_id: departmentId }),
  });
}

export async function bulkDeleteDevices(deviceIds: string[]): Promise<BulkActionResponse> {
  return request<BulkActionResponse>('/devices/bulk/delete', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_ids: deviceIds }),
  });
}

export async function exportDevicesCSV(params: DeviceListParams = {}): Promise<void> {
  const q = buildDeviceQS(params);
  const res = await fetch(`/api/v1/devices/export${q ? `?${q}` : ''}`, {
    credentials: 'include',
  });
  if (!res.ok) throw new Error('Export failed');
  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition') ?? '';
  const match = disposition.match(/filename=(.+)/);
  const filename = match ? match[1] : 'devices.csv';
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}
