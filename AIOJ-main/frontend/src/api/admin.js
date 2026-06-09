import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const adminApi = {
  getUsers: () => USE_MOCK ? mockApi.getAdminUsers() : http.get('/admin/users'),
  updateUserRole: (id, data) => USE_MOCK ? mockApi.updateAdminUserRole(id, data) : http.put(`/admin/users/${id}/role`, data),
  getAuditLogs: params => USE_MOCK ? mockApi.getAuditLogs(params) : http.get('/admin/audit-logs', { params })
}
