// Device API calls.

import { request } from './client';
import type { DeviceListResponse, DeviceDetailResponse } from '../types';

export async function getDevices(hostname?: string, os?: string): Promise<DeviceListResponse> {
  const params = new URLSearchParams();
  if (hostname) params.set('hostname', hostname);
  if (os) params.set('os', os);
  const qs = params.toString();
  return request<DeviceListResponse>(`/devices${qs ? `?${qs}` : ''}`);
}

export async function getDevice(id: string): Promise<DeviceDetailResponse> {
  return request<DeviceDetailResponse>(`/devices/${id}`);
}
