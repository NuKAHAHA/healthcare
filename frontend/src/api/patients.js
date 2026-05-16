import api from './axios';

export const getPatients = (page = 1, pageSize = 50) =>
  api.get('/patients', { params: { page, pageSize } });

export const getPatient = (id) =>
  api.get(`/patients/${id}`);

export const createPatient = (data) =>
  api.post('/patients', data);
