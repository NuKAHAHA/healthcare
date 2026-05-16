import { NavLink } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const NAV = [
  { to: '/dashboard',    label: 'Dashboard',    icon: '🏠', roles: null },
  { to: '/patients',     label: 'Patients',     icon: '👤', roles: ['registrar', 'admin', 'doctor'] },
  { to: '/appointments', label: 'Appointments', icon: '📅', roles: ['registrar', 'admin', 'doctor'] },
  { to: '/reports',      label: 'Reports',      icon: '📋', roles: ['doctor', 'admin'] },
  { to: '/audit-logs',   label: 'Audit Logs',   icon: '🔍', roles: ['admin'] },
];

export default function Sidebar() {
  const { user, logout } = useAuth();

  const visible = NAV.filter((l) => !l.roles || l.roles.includes(user?.role));

  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <span className="brand-icon">+</span>
        <span className="brand-name">HealthCare</span>
      </div>

      <nav className="sidebar-nav">
        {visible.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            className={({ isActive }) =>
              'nav-link' + (isActive ? ' nav-link-active' : '')
            }
          >
            <span className="nav-link-icon">{l.icon}</span>
            {l.label}
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-user">
          <span className="user-name">{user?.firstName} {user?.lastName}</span>
          <span className={`badge badge-role badge-${user?.role}`}>{user?.role}</span>
        </div>
        <button className="btn-logout" onClick={logout}>
          🚪 Sign out
        </button>
      </div>
    </aside>
  );
}
