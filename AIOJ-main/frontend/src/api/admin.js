import http from './index'

export const adminApi = {
  getUsers: () => http.get('/admin/users'),
  updateUserRole: (id, data) => http.put(`/admin/users/${id}/role`, data),
  getAuditLogs: params => http.get('/admin/audit-logs', { params })
}
