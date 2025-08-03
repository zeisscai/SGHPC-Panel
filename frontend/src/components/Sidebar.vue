<template>
  <v-navigation-drawer
    app
    permanent
    width="260"
    color="sidebar"
    class="sidebar-container"
    v-if="$route.name !== 'Login'"
  >
    <v-list-item class="px-2 py-4">
      <v-list-item>
        <v-list-item-title class="text-h6 font-weight-bold">
          SGHPC Panel
        </v-list-item-title>
      </v-list-item>
    </v-list-item>
    
    <v-divider></v-divider>
    
    <v-list nav density="compact" class="mt-2">
      <v-list-item
        link
        to="/"
        :active="$route.path === '/'"
        class="sidebar-item"
        :ripple="false"
      >
        <template v-slot:prepend>
          <v-icon class="sidebar-icon">mdi-view-dashboard</v-icon>
        </template>
        <v-list-item-title>Overview</v-list-item-title>
      </v-list-item>
      
      <v-list-group value="system" class="list-group">
        <template v-slot:activator="{ props }">
          <v-list-item
            v-bind="props"
            :active="isSystemRoute"
            class="sidebar-item"
            :ripple="false"
          >
            <template v-slot:prepend>
              <v-icon class="sidebar-icon">mdi-cog</v-icon>
            </template>
            <v-list-item-title>System</v-list-item-title>
          </v-list-item>
        </template>
        
        <v-list-item
          link
          to="/system/terminal"
          :active="$route.path === '/system/terminal'"
          class="submenu-item"
          :ripple="false"
        >
          <template v-slot:prepend>
            <v-icon size="small" class="submenu-icon">mdi-console</v-icon>
          </template>
          <v-list-item-title>Terminal</v-list-item-title>
        </v-list-item>
        
        <v-list-item
          link
          to="/system/files"
          :active="$route.path === '/system/files'"
          class="submenu-item"
          :ripple="false"
        >
          <template v-slot:prepend>
            <v-icon size="small" class="submenu-icon">mdi-file-document-multiple</v-icon>
          </template>
          <v-list-item-title>Files</v-list-item-title>
        </v-list-item>
      </v-list-group>
      
      <v-list-item
        link
        to="/settings"
        :active="$route.path === '/settings'"
        class="sidebar-item"
        :ripple="false"
      >
        <template v-slot:prepend>
          <v-icon class="sidebar-icon">mdi-wrench</v-icon>
        </template>
        <v-list-item-title>Settings</v-list-item-title>
      </v-list-item>
    </v-list>
    
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
</template>

<script>
import { computed } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'Sidebar',
  setup() {
    const router = useRouter()
    
    const isSystemRoute = computed(() => {
      return router.currentRoute.value.path.startsWith('/system')
    })
    
    const logout = () => {
      // 清除认证信息
      localStorage.removeItem('authToken')
      localStorage.removeItem('user')
      localStorage.removeItem('lastActivity')
      localStorage.removeItem('shouldChangePassword')
      
      // 跳转到登录页面
      router.push('/login')
    }
    
    return {
      isSystemRoute,
      logout
    }
  }
}
</script>

<style scoped>
.sidebar-container {
  border-right: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.sidebar-item,
.submenu-item {
  margin: 4px 8px;
  border-radius: 8px !important;
  transition: all 0.3s ease;
}

.sidebar-item:hover,
.submenu-item:hover {
  background-color: rgba(255, 255, 255, 0.1) !important;
}

.sidebar-item.v-list-item--active,
.submenu-item.v-list-item--active {
  background-color: rgba(25, 118, 210, 0.2) !important;
  color: #1976D2;
}

.sidebar-icon,
.submenu-icon {
  margin-right: 16px !important;
  background-color: transparent !important;
}

.v-theme--dark .sidebar-item.v-list-item--active,
.v-theme--dark .submenu-item.v-list-item--active {
  background-color: rgba(25, 118, 210, 0.3) !important;
  color: #64B5F6 !important;
}

.v-theme--dark .sidebar-item.v-list-item--active .sidebar-icon,
.v-theme--dark .submenu-item.v-list-item--active .submenu-icon {
  color: #64B5F6 !important;
}

/* 修复折叠箭头的黑色背景问题 */
.list-group :deep(.v-list-group__header .v-icon) {
  background-color: transparent !important;
}

/* 确保所有图标都没有黑色背景 */
:deep(.v-icon) {
  background-color: transparent !important;
}
</style>