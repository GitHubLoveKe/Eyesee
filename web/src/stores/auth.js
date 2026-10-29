import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/modules'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const refreshToken = ref('')
  const user = ref({ user_id: '', username: '', role: '' })

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value.username || '')
  const role = computed(() => user.value.role || '')

  const loadFromStorage = () => {
    token.value = localStorage.getItem('dcs_token') || ''
    refreshToken.value = localStorage.getItem('dcs_refresh_token') || ''
    const stored = localStorage.getItem('dcs_user')
    if (stored) {
      try {
        user.value = JSON.parse(stored)
      } catch {
        user.value = { user_id: '', username: '', role: '' }
      }
    }
  }

  const login = async (loginUsername, password) => {
    const res = await authApi.login({ username: loginUsername, password })
    const data = res.data
    token.value = data.token
    refreshToken.value = data.refresh_token || ''
    user.value = {
      user_id: data.user_id || '',
      username: data.username || loginUsername,
      role: data.role || '',
    }
    localStorage.setItem('dcs_token', data.token)
    if (data.refresh_token) {
      localStorage.setItem('dcs_refresh_token', data.refresh_token)
    }
    localStorage.setItem('dcs_user', JSON.stringify(user.value))
    return data
  }

  const doRefreshToken = async () => {
    if (!refreshToken.value) {
      logout()
      return false
    }
    try {
      const res = await authApi.refresh({ refresh_token: refreshToken.value })
      const data = res.data
      token.value = data.token
      localStorage.setItem('dcs_token', data.token)
      if (data.refresh_token) {
        refreshToken.value = data.refresh_token
        localStorage.setItem('dcs_refresh_token', data.refresh_token)
      }
      return true
    } catch {
      logout()
      return false
    }
  }

  const logout = () => {
    token.value = ''
    refreshToken.value = ''
    user.value = { user_id: '', username: '', role: '' }
    localStorage.removeItem('dcs_token')
    localStorage.removeItem('dcs_refresh_token')
    localStorage.removeItem('dcs_user')
  }

  return {
    token,
    refreshToken,
    user,
    isLoggedIn,
    username,
    role,
    loadFromStorage,
    login,
    doRefreshToken,
    logout,
  }
})
