import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const studyPlanApi = {
  getList: () => USE_MOCK ? mockApi.getStudyPlans() : http.get('/study-plans'),
  getDetail: id => USE_MOCK ? mockApi.getStudyPlanDetail(id) : http.get(`/study-plans/${id}`),
  getDailyChallenge: () => USE_MOCK ? mockApi.getDailyChallenge() : http.get('/daily-challenge'),
  getCheckins: () => USE_MOCK ? mockApi.getStudyCheckins() : http.get('/study-plans/checkins')
}
