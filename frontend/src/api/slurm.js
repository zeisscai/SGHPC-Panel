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

// 获取Slurm配置
export async function fetchSlurmConfig() {
  try {
    const response = await apiClient.get('/slurm/config')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch Slurm configuration')
  }
}

// 更新Slurm配置
export async function updateSlurmConfig(config) {
  try {
    const response = await apiClient.post('/slurm/config/update', config)
    return response.data
  } catch (error) {
    throw new Error('Failed to update Slurm configuration')
  }
}

// 获取部署状态
export async function fetchDeploymentStatus() {
  try {
    const response = await apiClient.get('/slurm/deploy/status')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch deployment status')
  }
}

// 启动部署
export async function startDeployment() {
  try {
    const response = await apiClient.post('/slurm/deploy/start')
    return response.data
  } catch (error) {
    throw new Error('Failed to start deployment')
  }
}

// 获取部署日志
export async function fetchDeploymentLogs() {
  try {
    const response = await apiClient.get('/slurm/deploy/logs')
    return response.data
  } catch (error) {
    throw new Error('Failed to fetch deployment logs')
  }
}

// 控制Slurm服务
export async function controlSlurmService(action) {
  try {
    const response = await apiClient.post('/slurm/service/control', { action })
    return response.data
  } catch (error) {
    throw new Error(`Failed to ${action} Slurm service`)
  }
}