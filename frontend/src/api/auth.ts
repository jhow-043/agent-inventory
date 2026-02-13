// Authentication API calls.

import { request } from './client';

export async function login(username: string, password: string): Promise<void> {
  await request<{ message: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function logout(): Promise<void> {
  await request<{ message: string }>('/auth/logout', {
    method: 'POST',
  });
}
