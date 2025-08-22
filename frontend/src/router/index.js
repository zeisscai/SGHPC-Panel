import { createRouter, createWebHistory } from 'vue-router'
import Overview from '../views/Overview.vue'
import Terminal from '../views/Terminal.vue'
import FileManagement from '../views/FileManagement.vue'
import Login from '../views/Login.vue'
import Spack from '../views/Spack.vue'
import User from '../views/User.vue'

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
    path: '/system/terminal',
    name: 'Terminal',
    component: Terminal,
    meta: { requiresAuth: true }
  },
  {
    path: '/system/files',
    name: 'FileManagement',
    component: FileManagement,
    meta: { requiresAuth: true }
  },
  {
    path: '/system/spack',
    name: 'Spack',
    component: Spack,
    meta: { requiresAuth: true }
  },
  {
    path: '/system/user',
    name: 'User',
    component: User,
    meta: { requiresAuth: true }
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
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    } else {
      next()
    }
  } else if (to.matched.some(record => record.meta.requiresGuest) && isAuthenticated) {
    // 如果已认证且访问的是访客专用页面，则重定向到主页
    next({ path: '/' })
  } else {
    next()
  }
})

export default router