// Dashboard API client functions.

import { request } from './client';
import type { DashboardStats } from '../types';

export async function getStats(): Promise<DashboardStats> {
  return request<DashboardStats>('/dashboard/stats');
}
