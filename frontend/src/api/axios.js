import axios from 'axios';
import { showToast } from './toastBridge';

// Access token in module memory — survives React re-renders, cleared on refresh.
// Page refresh restores it via the httpOnly cookie → /auth/refresh flow.
let _accessToken = null;

export const setToken  = (t) => { _accessToken = t; };
export const getToken  = ()  => _accessToken;
export const clearToken = () => { _accessToken = null; };

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // sends httpOnly refresh-token cookie automatically
});

// Attach access token to every request
api.interceptors.request.use((config) => {
  if (_accessToken) config.headers.Authorization = `Bearer ${_accessToken}`;
  return config;
});

// Auto-refresh on 401, toast on 403
let isRefreshing = false;
let refreshQueue = [];

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config;
    const status   = error.response?.status;

    if (status === 403) {
      showToast('Access denied — you do not have permission for this action', 'error');
      return Promise.reject(error);
    }

    if (status === 401 && !original._retry) {
      original._retry = true;

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          refreshQueue.push({ resolve, reject });
        }).then((token) => {
          original.headers.Authorization = `Bearer ${token}`;
          return api(original);
        });
      }

      isRefreshing = true;
      try {
        const { data } = await axios.post('/api/auth/refresh', {}, { withCredentials: true });
        _accessToken = data.accessToken;
        refreshQueue.forEach(({ resolve }) => resolve(_accessToken));
        refreshQueue = [];
        original.headers.Authorization = `Bearer ${_accessToken}`;
        return api(original);
      } catch (refreshErr) {
        refreshQueue.forEach(({ reject }) => reject(refreshErr));
        refreshQueue = [];
        _accessToken = null;
        window.location.href = '/login';
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default api;
