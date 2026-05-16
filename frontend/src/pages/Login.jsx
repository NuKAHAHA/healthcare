import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Login() {
  const [email, setEmail]               = useState('');
  const [password, setPassword]         = useState('');
  const [busy, setBusy]                 = useState(false);
  const [error, setError]               = useState('');
  const [lockoutSecs, setLockoutSecs]   = useState(0);

  const { login }  = useAuth();
  const navigate   = useNavigate();

  // Countdown: decrement lockoutSecs every second until it reaches 0
  useEffect(() => {
    if (lockoutSecs <= 0) return;
    const t = setTimeout(() => setLockoutSecs(s => s - 1), 1000);
    return () => clearTimeout(t);
  }, [lockoutSecs]);

  async function handleSubmit(e) {
    e.preventDefault();
    if (lockoutSecs > 0) return;
    setError('');
    setBusy(true);
    try {
      await login(email, password);
      navigate('/dashboard', { replace: true });
    } catch (err) {
      const status = err.response?.status;
      const data   = err.response?.data ?? {};

      if (status === 429) {
        const secs = typeof data.retry_after === 'number' ? data.retry_after : 900;
        setLockoutSecs(secs);
        setError('');
      } else if (status === 401) {
        const remaining = typeof data.attempts_remaining === 'number'
          ? data.attempts_remaining
          : null;
        if (remaining !== null && remaining <= 2) {
          const warn = remaining === 0
            ? 'Next failed attempt will lock your account.'
            : `${remaining} attempt${remaining === 1 ? '' : 's'} remaining before lockout.`;
          setError(`Invalid email or password. ${warn}`);
        } else {
          setError('Invalid email or password.');
        }
      } else if (!err.response) {
        setError('Cannot connect to server. Please check your connection.');
      } else {
        setError(data.message || 'Login failed. Please try again.');
      }
    } finally {
      setBusy(false);
    }
  }

  const fmtTime = (s) =>
    `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;

  const isLocked = lockoutSecs > 0;

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-header">
          <span className="brand-icon large">+</span>
          <h1>HealthCare Portal</h1>
          <p>Sign in to your account</p>
        </div>

        <form onSubmit={handleSubmit} className="login-form">
          {isLocked && (
            <div className="login-error">
              Account locked. Try again in <strong>{fmtTime(lockoutSecs)}</strong>
            </div>
          )}
          {!isLocked && error && (
            <div className="login-error">{error}</div>
          )}

          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoFocus
              placeholder="you@example.com"
              disabled={isLocked}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="••••••••"
              disabled={isLocked}
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-block"
            disabled={busy || isLocked}
          >
            {isLocked
              ? `Locked (${fmtTime(lockoutSecs)})`
              : busy
              ? 'Signing in…'
              : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
