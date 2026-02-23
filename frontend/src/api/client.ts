// Centralized API client with error handling.

const API_BASE = '/api/v1';

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    if (res.status === 401) {
      // Session expired — clear local flag and redirect to login
      localStorage.removeItem('authenticated');
      window.location.href = '/login';
      throw new ApiError('Unauthorized', 401);
    }
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(body.error || res.statusText, res.status);
  }

  // Handle 204 No Content and other empty responses
  if (res.status === 204 || res.headers.get('content-length') === '0') {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export { request, ApiError };
