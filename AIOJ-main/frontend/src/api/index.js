import axios from 'axios'
import { ElMessage } from 'element-plus'

const USE_MOCK = true

const http = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

http.interceptors.request.use(config => {
  const token = localStorage.getItem('toj_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  response => response.data,
  error => {
    const msg = error.response?.data?.message || '请求失败，请稍后重试'
    if (error.response?.status === 401) {
      localStorage.removeItem('toj_token')
      localStorage.removeItem('toj_user')
      window.location.href = '/login'
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(error)
  }
)

export { USE_MOCK }
export default http
