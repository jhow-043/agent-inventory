// Authentication context for managing login state across the app.

import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';
import * as authApi from '../api/auth';

interface AuthContextType {
  isAuthenticated: boolean;
  username: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  // Assume authenticated if the session cookie exists (httpOnly, so we can't read it).
  // The API will return 401 if the session is invalid.
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    return localStorage.getItem('authenticated') === 'true';
  });
  const [username, setUsername] = useState<string | null>(() => {
    return localStorage.getItem('username');
  });

  // Fetch current user info when authenticated.
  useEffect(() => {
    if (!isAuthenticated) return;
    authApi.getMe()
      .then((me) => {
        setUsername(me.username);
        localStorage.setItem('username', me.username);
      })
      .catch(() => {
        // 401 will auto-redirect via client.ts
      });
  }, [isAuthenticated]);

  const login = useCallback(async (user: string, password: string) => {
    await authApi.login(user, password);
    localStorage.setItem('authenticated', 'true');
    setIsAuthenticated(true);
    // Fetch username after login.
    try {
      const me = await authApi.getMe();
      setUsername(me.username);
      localStorage.setItem('username', me.username);
    } catch { /* ignore */ }
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    localStorage.removeItem('authenticated');
    localStorage.removeItem('username');
    setIsAuthenticated(false);
    setUsername(null);
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated, username, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
