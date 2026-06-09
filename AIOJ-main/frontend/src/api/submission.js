import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const submissionApi = {
  submit: data => USE_MOCK ? mockApi.submitCode(data) : http.post('/submissions', data),
  getList: params => USE_MOCK ? mockApi.getSubmissions(params) : http.get('/submissions', { params }),
  getDetail: id => USE_MOCK ? mockApi.getSubmissionDetail(id) : http.get(`/submissions/${id}`),
  stream: id => `/api/submissions/${id}/stream`,
  getCases: id => USE_MOCK ? mockApi.getSubmissionCases(id) : http.get(`/submissions/${id}/cases`),
  getOutput: id => USE_MOCK ? mockApi.getSubmissionOutput(id) : http.get(`/submissions/${id}/output`)
}
