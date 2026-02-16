// Department API calls.

import { request } from './client';
import type { Department, DepartmentListResponse } from '../types';

export async function getDepartments(): Promise<DepartmentListResponse> {
  return request<DepartmentListResponse>('/departments');
}

export async function createDepartment(name: string): Promise<Department> {
  return request<Department>('/departments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
}

export async function updateDepartment(id: string, name: string): Promise<Department> {
  return request<Department>(`/departments/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
}

export async function deleteDepartment(id: string): Promise<void> {
  await request(`/departments/${id}`, { method: 'DELETE' });
}
