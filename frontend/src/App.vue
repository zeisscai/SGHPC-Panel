<template>
  <v-app>
    <Sidebar />
    
    <v-main>
      <v-container fluid>
```

App.vue
```vue
<<<<<<< SEARCH
    const showPasswordAlert = ref(false)
    
    // 检查是否需要显示密码修改提醒
    const checkPasswordAlert = () => {
      const shouldChangePassword = localStorage.getItem('shouldChangePassword')
      showPasswordAlert.value = shouldChangePassword === 'true'
    }
    
    const closePasswordAlert = () => {
      showPasswordAlert.value = false
      localStorage.removeItem('shouldChangePassword')
    }
    
    const goToSettings = () => {
      router.push('/settings')
      closePasswordAlert()
    }
```

App.vue
```vue
<<<<<<< SEARCH
      checkPasswordAlert()
      // 检查认证状态等
        
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
    const showPasswordAlert = ref(false)
    
    // 检查是否需要显示密码修改提醒
    const checkPasswordAlert = () => {
      const shouldChangePassword = localStorage.getItem('shouldChangePassword')
      showPasswordAlert.value = shouldChangePassword === 'true'
    }
    
    const closePasswordAlert = () => {
      showPasswordAlert.value = false
      localStorage.removeItem('shouldChangePassword')
    }
    
    const goToSettings = () => {
      router.push('/settings')
      closePasswordAlert()
    }
    
    onMounted(() => {
      checkPasswordAlert()
    })
    
    return {
      router,
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
    }
  },
  mounted() {
    // 如果用户已认证，启动自动登出检查
    if (localStorage.getItem('authToken')) {
      // 每分钟检查一次是否超时
      this.autoLogoutInterval = setInterval(() => {
        const lastActivity = localStorage.getItem('lastActivity')
        if (lastActivity) {
          const now = Date.now()
          const elapsed = now - parseInt(lastActivity)
          // 5分钟超时 (5 * 60 * 1000 = 300000ms)
          if (elapsed > 300000) {
            this.logout()
            alert('Session expired due to inactivity. Please log in again.')
          }
        }
      }, 60000) // 每分钟检查一次
    }
  },
  beforeUnmount() {
    // 清理定时器
    if (this.autoLogoutInterval) {
      clearInterval(this.autoLogoutInterval)
    }
  },
  watch: {
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
</style>