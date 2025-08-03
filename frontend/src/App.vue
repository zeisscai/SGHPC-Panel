<template>
  <v-app>
    <v-navigation-drawer app permanent v-if="$route.name !== 'Login'">
      <v-list-item class="px-2">
        <v-list-item-title class="text-h6">HPC Control Panel</v-list-item-title>
      </v-list-item>
      
      <v-divider></v-divider>
      
      <v-list nav dense>
        <v-list-item 
          link 
          to="/" 
          :class="{ 'v-list-item--active': $route.path === '/' }"
        >
          <template v-slot:prepend>
            <v-icon :color="$route.path === '/' ? 'primary' : ''" class="no-bg-icon">mdi-server</v-icon>
          </template>
          <v-list-item-title :class="{ 'primary--text': $route.path === '/' }">
            Overview
          </v-list-item-title>
        </v-list-item>
        
        <v-list-group :value="isSystemRoute" @click="toggleSystemGroup">
          <template v-slot:activator="{ props }">
            <v-list-item v-bind="props">
              <template v-slot:prepend>
                <v-icon :color="isSystemRoute ? 'primary' : ''" class="no-bg-icon">mdi-cog</v-icon>
              </template>
              <v-list-item-title :class="{ 'primary--text': isSystemRoute }">
                System
              </v-list-item-title>
            </v-list-item>
          </template>
          
          <v-list-item 
            link 
            to="/system/terminal" 
            :class="{ 'v-list-item--active': $route.path === '/system/terminal' }"
          >
            <template v-slot:prepend>
              <v-icon :color="$route.path === '/system/terminal' ? 'primary' : ''" size="small" class="sidebar-icon submenu-icon">mdi-console</v-icon>
            </template>
            <v-list-item-title :class="{ 'primary--text': $route.path === '/system/terminal' }">
              Terminal
            </v-list-item-title>
          </v-list-item>
          
          <v-list-item 
            link 
            to="/system/files" 
            :class="{ 'v-list-item--active': $route.path === '/system/files' }"
          >
            <template v-slot:prepend>
              <v-icon :color="$route.path === '/system/files' ? 'primary' : ''" size="small" class="sidebar-icon submenu-icon">mdi-file-document-multiple</v-icon>
            </template>
            <v-list-item-title :class="{ 'primary--text': $route.path === '/system/files' }">
              File Management
            </v-list-item-title>
          </v-list-item>
        </v-list-group>
        
        <v-list-item 
          link 
          to="/settings" 
          :class="{ 'v-list-item--active': $route.path === '/settings' }"
        >
          <template v-slot:prepend>
            <v-icon :color="$route.path === '/settings' ? 'primary' : ''" class="sidebar-icon">mdi-wrench</v-icon>
          </template>
          <v-list-item-title :class="{ 'primary--text': $route.path === '/settings' }">
            Settings
          </v-list-item-title>
        </v-list-item>
      </v-list>
      
      <!-- 底部登出按钮 -->
      <template v-slot:append>
        <div class="pa-2">
          <v-btn 
            block 
            color="error" 
            @click="logout"
          >
            <v-icon start>mdi-logout</v-icon>
            Logout
          </v-btn>
        </div>
      </template>
    </v-navigation-drawer>

    <v-main>
      <v-container fluid>
        <v-alert
          v-if="showPasswordAlert"
          type="warning"
          dismissible
          class="ma-4"
          @input="closePasswordAlert"
        >
          <div class="d-flex align-center">
            <div class="flex-grow-1">
              <div class="font-weight-bold">Security Warning</div>
              <div>You are using the default password. Please change it immediately for security reasons.</div>
            </div>
            <v-btn
              color="white"
              text
              @click="goToSettings"
            >
              Change Password
            </v-btn>
          </div>
        </v-alert>
        
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

export default {
  name: 'App',
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
      showPasswordAlert,
      checkPasswordAlert,
      closePasswordAlert,
      goToSettings
    }
  },
  data: () => ({
    systemGroup: false
  }),
  computed: {
    isSystemRoute() {
      return this.$route.path.startsWith('/system');
    }
  },
  methods: {
    toggleSystemGroup() {
      // 修复需要点击两次的问题，不再手动控制systemGroup状态
    },
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
    isSystemRoute(newVal) {
      // 不再手动控制systemGroup状态
    },
    // 监听路由变化，检查是否需要显示密码提醒
    '$route'() {
      this.$nextTick(() => {
        this.checkPasswordAlert()
      })
    }
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

.v-list-group--active > .v-list-group__header {
  background-color: rgba(25, 118, 210, 0.1) !important;
}

/* 侧边栏图标样式 */
.sidebar-icon {
  background-color: transparent !important;
  border-radius: 0 !important;
  margin-right: 16px !important;
}

/* 子菜单项样式 */
.submenu-item {
  padding-left: 40px !important; /* 与父级菜单对齐 */
}

.submenu-item .v-list-item__prepend {
  margin-right: 16px !important;
}

/* 确保所有列表项的一致性 */
.v-list-item {
  display: flex !important;
  align-items: center !important;
  padding-left: 16px !important;
  padding-right: 16px !important;
}

/* 确保子菜单项没有额外缩进 */
.v-list-group--items {
  padding-inline-start: 0 !important;
  padding-left: 0 !important;
}

.v-list-group--items .v-list-item {
  padding-left: 0 !important;
  padding-inline-start: 0 !important;
}

/* 确保prepend区域正确对齐 */
.v-list-item__prepend {
  margin-right: 16px !important;
  align-self: center !important;
}

/* 移除prepend区域的spacer */
.v-list-item__prepend .v-list-item__spacer {
  display: none !important;
}

.v-list-item__append .v-list-item__spacer {
  display: none !important;
}

/* 确保子菜单图标正确对齐 */
.submenu-icon {
  margin-left: 0 !important;
}

/* 折叠图标样式 */
.v-list-group__header .v-icon {
  background-color: transparent !important;
}

/* 悬停效果 */
.v-list-item:hover .sidebar-icon {
  background-color: transparent !important;
}

/* 确保文字对齐 */
.v-list-item-title {
  align-self: center !important;
  flex: 1 !important;
  padding-left: 0 !important;
}
</style>