// User management API calls.

import { request } from './client';
import type { UserListResponse } from '../types';

export async function getUsers(): Promise<UserListResponse> {
  return request<UserListResponse>('/users');
}

export async function createUser(username: string, password: string, role: string): Promise<{ message: string }> {
  return request<{ message: string }>('/users', {
    method: 'POST',
    body: JSON.stringify({ username, password, role }),
  });
}

export async function updateUser(id: string, data: { username?: string; password?: string; role?: string }): Promise<{ message: string }> {
  return request<{ message: string }>(`/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function deleteUser(id: string): Promise<{ message: string }> {
  return request<{ message: string }>(`/users/${id}`, {
    method: 'DELETE',
  });
}
