import http from './index'

export const studyPlanApi = {
  getList: () => http.get('/study-plans'),
  getDetail: id => http.get(`/study-plans/${id}`),
  getDailyChallenge: () => http.get('/daily-challenge'),
  getCheckins: () => http.get('/study-plans/checkins')
}
