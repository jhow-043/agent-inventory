// Authentication context for managing login state across the app.

import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import * as authApi from '../api/auth';

interface AuthContextType {
  isAuthenticated: boolean;
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

  const login = useCallback(async (username: string, password: string) => {
    await authApi.login(username, password);
    localStorage.setItem('authenticated', 'true');
    setIsAuthenticated(true);
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    localStorage.removeItem('authenticated');
    setIsAuthenticated(false);
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
