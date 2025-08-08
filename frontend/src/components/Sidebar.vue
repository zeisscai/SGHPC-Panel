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
        <v-list-item-title v-if="!isRail">Overview</v-list-item-title>
        <v-tooltip
          v-if="isRail"
          activator="parent"
          location="right"
        >Overview</v-tooltip>
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
          <v-list-item-title v-if="!isRail">Terminal</v-list-item-title>
          <v-tooltip
            v-if="isRail"
            activator="parent"
            location="right"
          >Terminal</v-tooltip>
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
          <v-list-item-title v-if="!isRail">File Management</v-list-item-title>
          <v-tooltip
            v-if="isRail"
            activator="parent"
            location="right"
          >File Management</v-tooltip>
        </v-list-item>
      </v-list-group>
    </v-list>
    
    <template v-slot:append>
      <div class="pa-2">
        <v-btn
          block
          color="error"
          @click="logout"
          :variant="isRail ? 'icon' : 'flat'"
        >
          <v-icon :start="!isRail">mdi-logout</v-icon>
          <span v-if="!isRail">Logout</span>
          <v-tooltip
            v-if="isRail"
            activator="parent"
            location="right"
          >Logout</v-tooltip>
        </v-btn>
        
        <div v-if="isRail" class="d-flex justify-center mt-2">
          <v-btn 
            icon 
            @click="toggleRail"
            size="small"
            variant="text"
          >
            <v-icon>mdi-chevron-double-right</v-icon>
          </v-btn>
        </div>
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
    
    const isRail = computed(() => {
      // 这里可以根据需要实现侧边栏收缩逻辑
      return false
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
    
    const toggleRail = () => {
      // 实现侧边栏收缩逻辑
    }
    
    return {
      isSystemRoute,
      isRail,
      logout,
      toggleRail
    }
  }
}
</script>

<style scoped>
.sidebar-container {
  z-index: 999;
}

.sidebar-item {
  margin-bottom: 4px;
  border-radius: 4px;
  margin-left: 8px;
  margin-right: 8px;
}

.sidebar-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.sidebar-icon {
  margin-right: 12px;
}

.submenu-item {
  margin-bottom: 4px;
  border-radius: 4px;
  margin-left: 16px;
  margin-right: 8px;
  padding-left: 56px !important;
}

.submenu-icon {
  margin-right: 12px;
}

.list-group :deep(.v-list-group__items .v-list-item) {
  padding-inline-start: 16px !important;
}
</style>