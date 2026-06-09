import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const problemApi = {
  getList: params => USE_MOCK ? mockApi.getProblems(params) : http.get('/problems', { params }),
  getDetail: id => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/problems/${id}`),
  favorite: id => USE_MOCK ? mockApi.favoriteProblem(id) : http.post(`/problems/${id}/favorite`),
  unfavorite: id => USE_MOCK ? mockApi.unfavoriteProblem(id) : http.delete(`/problems/${id}/favorite`),
  saveSolution: (id, data) => USE_MOCK ? mockApi.saveProblemSolution(id, data) : http.post(`/problems/${id}/solution`, data),
  getMySolutions: () => USE_MOCK ? mockApi.getMySolutions() : http.get('/my/solutions'),
  getMySolutionDetail: id => USE_MOCK ? mockApi.getMySolutionDetail(id) : http.get(`/my/solutions/${id}`),
  getSolutionDetail: id => USE_MOCK ? mockApi.getSolutionDetail(id) : http.get(`/solutions/${id}`),
  runCode: (id, data) => USE_MOCK ? mockApi.runProblemCode(id, data) : http.post(`/problems/${id}/run`, data),
  getAnnouncements: () => USE_MOCK ? mockApi.getAnnouncements() : http.get('/announcements'),
  create: data => USE_MOCK ? mockApi.createProblem(data) : http.post('/problems', data),
  getAdminDetail: id => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/admin/problems/${id}`),
  getVersions: id => USE_MOCK ? mockApi.getProblemVersions(id) : http.get(`/admin/problems/${id}/versions`),
  publish: (id, data = {}) => USE_MOCK ? mockApi.publishProblem(id, data) : http.post(`/admin/problems/${id}/publish`, data),
  rollback: (id, data) => USE_MOCK ? mockApi.rollbackProblem(id, data) : http.post(`/admin/problems/${id}/rollback`, data),
  rejudge: (id, data = {}) => USE_MOCK ? mockApi.rejudgeProblem(id, data) : http.post(`/admin/problems/${id}/rejudge`, data),
  getRejudgeJobs: id => USE_MOCK ? mockApi.getRejudgeJobs(id) : http.get(`/admin/problems/${id}/rejudge-jobs`),
  update: (id, data) => USE_MOCK ? mockApi.updateProblem(id, data) : http.put(`/problems/${id}`, data),
  remove: id => USE_MOCK ? mockApi.deleteProblem(id) : http.delete(`/problems/${id}`)
}
