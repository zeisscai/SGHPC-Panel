<template>
  <v-app>
    <Sidebar />
    
    <v-main>
      <v-container fluid>
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Sidebar from './components/Sidebar.vue'

export default {
  name: 'App',
  components: {
    Sidebar
  },
  setup() {
    const router = useRouter()
    
    onMounted(() => {
      // 检查认证状态等
    })
    
    return {
      router
    }
  },
  data: () => ({
    autoLogoutInterval: null
  }),
  methods: {
    logout() {
      // 清除认证信息
      localStorage.removeItem('authToken')
      localStorage.removeItem('user')
      localStorage.removeItem('lastActivity')
      localStorage.removeItem('shouldChangePassword')
      
      // 跳转到登录页面
      this.router.push('/login')
    },
    updateLastActivity() {
      // 更新最后活动时间
      localStorage.setItem('lastActivity', Date.now().toString())
    }
  },
  mounted() {
    // 如果用户已认证，启动自动登出检查
    if (localStorage.getItem('authToken')) {
      // 更新最后活动时间
      this.updateLastActivity()
      
      // 监听用户活动事件
      const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
      events.forEach(event => {
        document.addEventListener(event, this.updateLastActivity, true)
      })
      
      // 每分钟检查一次是否超时
      this.autoLogoutInterval = setInterval(() => {
        const lastActivity = localStorage.getItem('lastActivity')
        if (lastActivity) {
          const now = Date.now()
          const elapsed = now - parseInt(lastActivity)
          // 30分钟超时 (30 * 60 * 1000 = 1800000ms)
          if (elapsed > 1800000) {
            this.logout()
            alert('Session expired due to inactivity. Please log in again.')
          }
        }
      }, 60000) // 每分钟检查一次
    }
  },
  beforeUnmount() {
    // 清理定时器和事件监听器
    if (this.autoLogoutInterval) {
      clearInterval(this.autoLogoutInterval)
    }
    
    // 移除事件监听器
    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
    events.forEach(event => {
      document.removeEventListener(event, this.updateLastActivity, true)
    })
  }
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.v-application {
  font-family: 'Roboto', sans-serif !important;
}

.v-theme--dark {
  background-color: #121212;
  color: #ffffff;
}

.v-list-item--active {
  background-color: rgba(25, 118, 210, 0.1) !important;
}
