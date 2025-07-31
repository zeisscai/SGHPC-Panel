import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    // 指定入口文件
    rollupOptions: {
      input: resolve(__dirname, 'web/static/index.html')
    },
    // 指定输出目录
    outDir: resolve(__dirname, 'web/static/dist'),
    // 清空输出目录
    emptyOutDir: true
  },
  root: resolve(__dirname, 'web/static'),
  base: '/static/dist/',
  resolve: {
    alias: {
      '@': resolve(__dirname, 'web/static')
    }
  },
  server: {
    // 添加服务器配置，便于调试
    host: '0.0.0.0',
    port: 3000
  }
})