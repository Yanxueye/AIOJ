import http from './index'

export const submissionApi = {
  submit: data => http.post('/submissions', data),
  getList: params => http.get('/submissions', { params }),
  getDetail: id => http.get(`/submissions/${id}`),
  stream: id => `/api/submissions/${id}/stream`,
  getCases: id => http.get(`/submissions/${id}/cases`),
  getOutput: id => http.get(`/submissions/${id}/output`)
}
