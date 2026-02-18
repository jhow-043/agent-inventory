// Audit logs API client functions.

import { request } from './client';

export interface AuditLog {
  id: string;
  user_id: string | null;
  username: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface AuditLogListResponse {
  logs: AuditLog[];
  total: number;
}

export interface AuditLogParams {
  user_id?: string;
  action?: string;
  resource_type?: string;
  resource_id?: string;
  limit?: number;
  offset?: number;
}

export async function getAuditLogs(params: AuditLogParams = {}): Promise<AuditLogListResponse> {
  const qs = new URLSearchParams();
  if (params.user_id) qs.set('user_id', params.user_id);
  if (params.action) qs.set('action', params.action);
  if (params.resource_type) qs.set('resource_type', params.resource_type);
  if (params.resource_id) qs.set('resource_id', params.resource_id);
  if (params.limit) qs.set('limit', String(params.limit));
  if (params.offset) qs.set('offset', String(params.offset));
  const q = qs.toString();
  return request<AuditLogListResponse>(`/audit-logs${q ? `?${q}` : ''}`);
}
