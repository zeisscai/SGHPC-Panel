// 格式化日期时间
export function formatDateTime(date) {
  if (!date) return ''
  
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 格式化CPU和内存使用率
export function formatPercentage(value) {
  if (typeof value !== 'number') return '0.00%'
  return `${value.toFixed(2)}%`
}

// 格式化等待时间和计算时间
export function formatDuration(duration) {
  if (!duration) return '0m'
  return duration
}

export default {
  formatDateTime,
  formatPercentage,
  formatDuration
}