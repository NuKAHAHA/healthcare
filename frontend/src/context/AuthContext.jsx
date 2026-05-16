import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { login as apiLogin, logout as apiLogout, refresh as apiRefresh } from '../api/auth';
import { setToken, clearToken } from '../api/axios';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser]       = useState(null);
  const [loading, setLoading] = useState(true);

  // Restore session on mount by attempting a silent refresh
  useEffect(() => {
    apiRefresh()
      .then(({ data }) => {
        setToken(data.accessToken);
        setUser(data.user);
      })
      .catch(() => {
        // No valid refresh token — stay logged out
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email, password) => {
    const { data } = await apiLogin(email, password);
    setToken(data.accessToken);
    setUser(data.user);
    return data.user;
  }, []);

  const logout = useCallback(async () => {
    try { await apiLogout(); } catch { /* best-effort */ }
    clearToken();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
