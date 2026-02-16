// Device API calls.

import { request } from './client';
import type { DeviceListResponse, DeviceDetailResponse } from '../types';

export interface DeviceListParams {
  hostname?: string;
  os?: string;
  status?: string;
  sort?: string;
  order?: string;
  page?: number;
  limit?: number;
}

export async function getDevices(params: DeviceListParams = {}): Promise<DeviceListResponse> {
  const qs = new URLSearchParams();
  if (params.hostname) qs.set('hostname', params.hostname);
  if (params.os) qs.set('os', params.os);
  if (params.status) qs.set('status', params.status);
  if (params.sort) qs.set('sort', params.sort);
  if (params.order) qs.set('order', params.order);
  if (params.page) qs.set('page', String(params.page));
  if (params.limit) qs.set('limit', String(params.limit));
  const q = qs.toString();
  return request<DeviceListResponse>(`/devices${q ? `?${q}` : ''}`);
}

export async function getDevice(id: string): Promise<DeviceDetailResponse> {
  return request<DeviceDetailResponse>(`/devices/${id}`);
}
