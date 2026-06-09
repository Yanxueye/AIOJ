import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const userApi = {
  login: data => USE_MOCK ? mockApi.login(data) : http.post('/auth/login', data),
  register: data => USE_MOCK ? mockApi.register(data) : http.post('/auth/register', data),
  getProfile: () => USE_MOCK ? mockApi.getProfile() : http.get('/user/profile'),
  updateProfile: data => USE_MOCK ? mockApi.updateProfile(data) : http.put('/user/profile', data)
}
