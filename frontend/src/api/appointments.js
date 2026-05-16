import api from './axios';

export const getAppointments = (page = 1, pageSize = 50) =>
  api.get('/appointments', { params: { page, pageSize } });

export const createAppointment = (data) =>
  api.post('/appointments', data);

export const updateAppointmentStatus = (id, status) =>
  api.patch(`/appointments/${id}/status`, { status });
