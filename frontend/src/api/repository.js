import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000, // 30秒超时，因为某些操作可能需要较长时间
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理通用错误
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response?.status === 401) {
      // 未授权，跳转到登录页
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

/**
 * 清理软件包缓存
 * 执行 dnf clean all 和 dnf makecache 命令
 */
export const cleanRepositoryCache = async () => {
  try {
    const response = await api.post('/api/repository/clean-cache')
    return response
  } catch (error) {
    console.error('清理缓存失败:', error)
    throw error
  }
}

/**
 * 切换软件源
 * @param {string} type - 软件源类型 ('default' | 'ustc')
 */
export const switchRepository = async (type) => {
  try {
    const response = await api.post('/api/repository/switch', { type })
    return response
  } catch (error) {
    console.error('切换软件源失败:', error)
    throw error
  }
}

/**
 * 获取当前软件源状态
 * 返回当前使用的软件源类型、URL和可访问性状态
 */
export const getRepositoryStatus = async () => {
  try {
    const response = await api.get('/api/repository/status')
    return response
  } catch (error) {
    console.error('获取软件源状态失败:', error)
    throw error
  }
}

/**
 * 设置自定义软件源
 * @param {string} url - 自定义软件源URL
 */
export const setCustomRepository = async (url) => {
  try {
    const response = await api.post('/api/repository/custom', { url })
    return response
  } catch (error) {
    console.error('设置自定义软件源失败:', error)
    throw error
  }
}

/**
 * 测试软件源连接性
 * @param {string} url - 要测试的软件源URL
 */
export const testRepositoryConnection = async (url) => {
  try {
    const response = await api.post('/api/repository/test', { url })
    return response
  } catch (error) {
    console.error('测试软件源连接失败:', error)
    throw error
  }
}

/**
 * 获取可用的预设软件源列表
 */
export const getAvailableRepositories = async () => {
  try {
    const response = await api.get('/api/repository/available')
    return response
  } catch (error) {
    console.error('获取可用软件源列表失败:', error)
    throw error
  }
}

export default {
  cleanRepositoryCache,
  switchRepository,
  getRepositoryStatus,
  setCustomRepository,
  testRepositoryConnection,
  getAvailableRepositories
}