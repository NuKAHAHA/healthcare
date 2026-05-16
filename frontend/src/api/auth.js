import axios from 'axios';
import api from './axios';

// Raw axios (no interceptor) so 401/429 errors reach Login.jsx directly
// without triggering the token-refresh loop.
export const login = (email, password) =>
  axios.post('/api/auth/login', { email, password }, { withCredentials: true });

export const register = (data) =>
  api.post('/auth/register', data);

// Uses a raw axios call so it sends the httpOnly cookie without going
// through the interceptor (which would try to re-auth and loop).
export const refresh = () =>
  axios.post('/api/auth/refresh', {}, { withCredentials: true });

export const logout = () =>
  api.post('/auth/logout');
