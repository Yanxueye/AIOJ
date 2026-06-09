import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const problemApi = {
  getList: params => USE_MOCK ? mockApi.getProblems(params) : http.get('/problems', { params }),
  getDetail: id => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/problems/${id}`),
  getAnnouncements: () => USE_MOCK ? mockApi.getAnnouncements() : http.get('/announcements'),
  create: data => USE_MOCK ? mockApi.createProblem(data) : http.post('/problems', data),
  getAdminDetail: id => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/admin/problems/${id}`),
  update: (id, data) => USE_MOCK ? mockApi.updateProblem(id, data) : http.put(`/problems/${id}`, data),
  remove: id => USE_MOCK ? mockApi.deleteProblem(id) : http.delete(`/problems/${id}`)
}
