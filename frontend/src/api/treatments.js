import api from './axios';

export const createTreatment = (data) =>
  api.post('/treatments', data);

export const getReport = (appointmentId) =>
  api.get(`/reports/${appointmentId}`);
