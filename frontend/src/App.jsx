import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import Spinner from './components/Spinner';

import Login        from './pages/Login';
import Dashboard    from './pages/Dashboard';
import Patients     from './pages/Patients';
import Appointments from './pages/Appointments';
import Reports      from './pages/Reports';
import AuditLogs    from './pages/AuditLogs';
import Unauthorized from './pages/Unauthorized';

export default function App() {
  const { loading } = useAuth();
  if (loading) return <Spinner fullPage />;

  return (
    <Routes>
      <Route path="/login"        element={<Login />} />
      <Route path="/unauthorized" element={<Unauthorized />} />

      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />

          <Route element={<ProtectedRoute roles={['registrar', 'admin', 'doctor']} />}>
            <Route path="/patients" element={<Patients />} />
          </Route>

          <Route element={<ProtectedRoute roles={['registrar', 'admin', 'doctor']} />}>
            <Route path="/appointments" element={<Appointments />} />
          </Route>

          <Route element={<ProtectedRoute roles={['doctor', 'admin']} />}>
            <Route path="/reports" element={<Reports />} />
          </Route>

          <Route element={<ProtectedRoute roles={['admin']} />}>
            <Route path="/audit-logs" element={<AuditLogs />} />
          </Route>
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
