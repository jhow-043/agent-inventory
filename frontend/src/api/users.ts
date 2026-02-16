// User management API calls.

import { request } from './client';
import type { UserListResponse } from '../types';

export async function getUsers(): Promise<UserListResponse> {
  return request<UserListResponse>('/users');
}

export async function createUser(username: string, password: string): Promise<{ message: string }> {
  return request<{ message: string }>('/users', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function deleteUser(id: string): Promise<{ message: string }> {
  return request<{ message: string }>(`/users/${id}`, {
    method: 'DELETE',
  });
}
