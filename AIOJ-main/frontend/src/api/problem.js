import http from './index'

export const problemApi = {
  getList: params => http.get('/problems', { params }),
  getDetail: id => http.get(`/problems/${id}`),
  favorite: id => http.post(`/problems/${id}/favorite`),
  unfavorite: id => http.delete(`/problems/${id}/favorite`),
  saveSolution: (id, data) => http.post(`/problems/${id}/solution`, data),
  getMySolutionDetail: id => http.get(`/my/solutions/${id}`),
  likeSolution: sid => http.post(`/solutions/${sid}/like`),
  deleteSolution: sid => http.delete(`/solutions/${sid}`),
  runCode: (id, data) => http.post(`/problems/${id}/run`, data),
  getAnnouncements: () => http.get('/announcements'),
  create: data => http.post('/problems', data),
  getAdminDetail: id => http.get(`/admin/problems/${id}`),
  getVersions: id => http.get(`/admin/problems/${id}/versions`),
  publish: (id, data = {}) => http.post(`/admin/problems/${id}/publish`, data),
  rollback: (id, data) => http.post(`/admin/problems/${id}/rollback`, data),
  update: (id, data) => http.put(`/problems/${id}`, data),
  remove: id => http.delete(`/problems/${id}`)
}
