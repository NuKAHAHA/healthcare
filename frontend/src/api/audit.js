import api from './axios';

export const getAuditLogs = (userId, page = 1, pageSize = 50) =>
  api.get(`/audit-logs/${userId}`, { params: { page, pageSize } });
