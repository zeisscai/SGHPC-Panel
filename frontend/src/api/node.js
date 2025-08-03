import axios from 'axios'

const API_BASE = '/api'

// 创建axios实例并添加请求拦截器
const apiClient = axios.create({
  baseURL: API_BASE
})

// 添加请求拦截器来包含认证token
apiClient.interceptors.request.use(
  config => {
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

export async function fetchManagementNode() {
  try {
    const response = await apiClient.get('/management-node')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch management node')
  }
}

export async function fetchComputeNodes() {
  try {
    const response = await apiClient.get('/compute-nodes')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch compute nodes')
  }
}

export async function fetchSlurmJobs() {
  try {
    const response = await apiClient.get('/slurm-jobs')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch SLURM jobs')
  }
}

// 登录API
export async function login(username, password) {
  try {
    const response = await apiClient.post('/login', { username, password })
    return response.data
  } catch (error) {
    throw new Error('Login failed')
  }
}

// 修改密码API
export async function changePassword(currentPassword, newPassword) {
  try {
    const response = await apiClient.post('/change-password', { 
      current_password: currentPassword, 
      new_password: newPassword 
    })
    return response.data
  } catch (error) {
    throw new Error('Failed to change password')
  }
}

// 上传文件
export const uploadFile = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return apiClient.post('/file/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 下载文件
export const downloadFile = (filePath) => {
  return apiClient.get('/file/download', {
    params: { path: filePath },
    responseType: 'blob'
  })
}

// 修改文件权限
export const changeFilePermissions = (filePath, permissions) => {
  return apiClient.post('/file/permissions', { 
    path: filePath,
    permissions: permissions
  })
}