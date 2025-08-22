<template>
  <div class="system">
    <div v-if="loading" class="loading-container">
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
      <p class="mt-2">加载中...</p>
    </div>
    <div v-else-if="error" class="error-container">
      <v-alert type="error" outlined>
        {{ error }}
      </v-alert>
    </div>
    <router-view v-else></router-view>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

export default {
  name: 'System',
  setup() {
    const route = useRoute()
    const loading = ref(false)
    const error = ref(null)

    onMounted(() => {
      // 组件挂载时的初始化逻辑
      console.log('System component mounted')
    })

    return {
      loading,
      error
    }
  }
}
</script>

<style scoped>
.system {
  animation: fadeIn 0.5s ease-in;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-container, .error-container {
  text-align: center;
  width: 100%;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>