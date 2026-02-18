import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './hooks/useAuth'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import DeviceList from './pages/DeviceList'
import DeviceDetail from './pages/DeviceDetail'
import Departments from './pages/Departments'
import DepartmentDetail from './pages/DepartmentDetail'
import Settings from './pages/Settings'
import AuditLogs from './pages/AuditLogs'

function App() {
  const { isAuthenticated } = useAuth()

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/login"
          element={isAuthenticated ? <Navigate to="/" replace /> : <Login />}
        />
        <Route
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<Dashboard />} />
          <Route path="/devices" element={<DeviceList />} />
          <Route path="/devices/:id" element={<DeviceDetail />} />
          <Route path="/departments" element={<Departments />} />
          <Route path="/departments/:id" element={<DepartmentDetail />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/audit-logs" element={<AuditLogs />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
