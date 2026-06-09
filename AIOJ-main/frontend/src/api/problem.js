import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const problemApi = {
  getList: params => USE_MOCK ? mockApi.getProblems(params) : http.get('/problems', { params }),
  getDetail: id => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/problems/${id}`),
  getAnnouncements: () => USE_MOCK ? mockApi.getAnnouncements() : http.get('/announcements')
}
