import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

let isRefreshing = false
let pendingRequests = []

const onRefreshed = (token) => {
  pendingRequests.forEach((cb) => cb(token))
  pendingRequests = []
}

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('dcs_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }
    return data
  },
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      if (isRefreshing) {
        return new Promise((resolve) => {
          pendingRequests.push((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(api(originalRequest))
          })
        })
      }
      isRefreshing = true
      const refreshTokenVal = localStorage.getItem('dcs_refresh_token')
      if (!refreshTokenVal) {
        localStorage.removeItem('dcs_token')
        localStorage.removeItem('dcs_refresh_token')
        localStorage.removeItem('dcs_user')
        router.push('/login')
        isRefreshing = false
        return Promise.reject(error)
      }
      try {
        const res = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshTokenVal })
        const data = res.data?.data || res.data
        const newToken = data.token
        localStorage.setItem('dcs_token', newToken)
        if (data.refresh_token) {
          localStorage.setItem('dcs_refresh_token', data.refresh_token)
        }
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        onRefreshed(newToken)
        return api(originalRequest)
      } catch {
        localStorage.removeItem('dcs_token')
        localStorage.removeItem('dcs_refresh_token')
        localStorage.removeItem('dcs_user')
        pendingRequests = []
        router.push('/login')
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }
    const msg = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

export default api
