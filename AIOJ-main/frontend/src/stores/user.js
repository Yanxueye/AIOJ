import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('toj_token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('toj_user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => userInfo.value?.username || '')
  const isAdmin = computed(() => userInfo.value?.role === 'admin')

  function setAuth(tokenVal, user) {
    token.value = tokenVal
    userInfo.value = user
    localStorage.setItem('toj_token', tokenVal)
    localStorage.setItem('toj_user', JSON.stringify(user))
  }

  function clearAuth() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('toj_token')
    localStorage.removeItem('toj_user')
  }

  async function login(credentials) {
    const res = await userApi.login(credentials)
    setAuth(res.data.token, res.data.user)
    return res
  }

  async function register(data) {
    const res = await userApi.register(data)
    return res
  }

  async function fetchProfile() {
    const res = await userApi.getProfile()
    userInfo.value = res.data
    localStorage.setItem('toj_user', JSON.stringify(res.data))
    return res.data
  }

  async function updateProfile(data) {
    const res = await userApi.updateProfile(data)
    userInfo.value = { ...userInfo.value, ...res.data }
    localStorage.setItem('toj_user', JSON.stringify(userInfo.value))
    return res.data
  }

  function logout() {
    clearAuth()
  }

  return {
    token, userInfo, isLoggedIn, isAdmin, username,
    login, register, logout, fetchProfile, updateProfile, setAuth
  }
})
