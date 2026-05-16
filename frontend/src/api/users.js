import api from './axios';

export const getUsers = (role) =>
  api.get('/users', { params: { role } });
