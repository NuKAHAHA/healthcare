import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const ROLE_DESC = {
  admin:     'Full system access — manage users, view audit logs, generate reports.',
  registrar: 'Register patients and schedule appointments.',
  doctor:    'View your appointments, record treatments and generate reports.',
};

export default function Dashboard() {
  const { user } = useAuth();

  return (
    <div className="page">
      <div className="page-header">
        <h2>Welcome back, {user?.firstName} {user?.lastName}</h2>
        <span className={`badge badge-role badge-${user?.role}`}>{user?.role}</span>
      </div>

      <p className="page-subtitle">{ROLE_DESC[user?.role]}</p>

      <div className="stat-grid">
        {(user?.role === 'registrar' || user?.role === 'admin') && (
          <StatCard label="Patients"     icon="👤" to="/patients" />
        )}
        {user?.role === 'doctor' && (
          <StatCard label="My Patients"  icon="👤" to="/patients" />
        )}
        {(user?.role === 'doctor' || user?.role === 'registrar' || user?.role === 'admin') && (
          <StatCard label="Appointments" icon="📅" to="/appointments" />
        )}
        {(user?.role === 'doctor' || user?.role === 'admin') && (
          <StatCard label="Reports"      icon="📋" to="/reports" />
        )}
        {user?.role === 'admin' && (
          <StatCard label="Audit Logs"   icon="🔍" to="/audit-logs" />
        )}
      </div>
    </div>
  );
}

function StatCard({ label, icon, to }) {
  return (
    <Link to={to} className="stat-card">
      <span className="stat-icon">{icon}</span>
      <span className="stat-label">{label}</span>
    </Link>
  );
}
