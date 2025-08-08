import axios from 'axios'

// 获取 Spack 状态
export const fetchSpackStatus = () => {
  return axios.get('/api/spack/status')
}

// 安装 Spack
export const installSpack = () => {
  return axios.post('/api/spack/install')
}

// 获取可安装的软件包列表
export const fetchAvailablePackages = () => {
  return axios.get('/api/spack/packages/available')
}

// 获取已安装的软件包列表
export const fetchInstalledPackages = () => {
  return axios.get('/api/spack/packages/installed')
}

// 安装软件包
export const installPackage = (packageName, options = '') => {
  return axios.post('/api/spack/package/install', {
    package_name: packageName,
    options: options
  })
}

// 卸载软件包
export const uninstallPackage = (packageName) => {
  return axios.post('/api/spack/package/uninstall', {
    package_name: packageName
  })
}

// 获取软件源配置
export const fetchRepositories = () => {
  return axios.get('/api/spack/repositories')
}

// 更新软件源配置
export const updateRepositories = (content) => {
  return axios.post('/api/spack/repositories/update', {
    content: content
  })
}