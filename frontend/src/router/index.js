import { createRouter, createWebHistory } from 'vue-router'
import Overview from '../views/Overview.vue'
import System from '../views/System.vue'
import Terminal from '../views/Terminal.vue'
import FileManagement from '../views/FileManagement.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/',
    name: 'Overview',
    component: Overview,
    meta: { requiresAuth: true }
  },
  {
    path: '/system',
    name: 'System',
    component: System,
    meta: { requiresAuth: true },
    children: [
      {
        path: 'terminal',
        name: 'Terminal',
        component: Terminal
      },
      {
        path: 'files',
        name: 'FileManagement',
        component: FileManagement
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 添加路由守卫
router.beforeEach((to, from, next) => {
  const isAuthenticated = !!localStorage.getItem('authToken')
  
  // 检查是否需要认证
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!isAuthenticated) {
      next('/login')
    } else {
      // 更新最后活动时间
      localStorage.setItem('lastActivity', Date.now().toString())
      next()
    }
  } 
  // 检查是否需要访客权限（如登录页面）
  else if (to.matched.some(record => record.meta.requiresGuest)) {
    if (isAuthenticated) {
      next('/')
    } else {
      next()
    }
  } 
  // 其他情况直接通过
  else {
    next()
  }
})

export default router