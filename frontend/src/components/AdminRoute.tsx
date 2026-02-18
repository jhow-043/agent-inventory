// AdminRoute restricts access to admin-only pages.
// Viewers are redirected to the dashboard.

import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import type { ReactNode } from 'react';

export default function AdminRoute({ children }: { children: ReactNode }) {
  const { role } = useAuth();

  if (role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}
