import http from './index'

export const userApi = {
  login: data => http.post('/auth/login', data),
  register: data => http.post('/auth/register', data),
  getProfile: () => http.get('/user/profile'),
  updateProfile: data => http.put('/user/profile', data),
  getRatingHistory: (limit = 100) => http.get('/user/rating-history', { params: { limit } })
}
