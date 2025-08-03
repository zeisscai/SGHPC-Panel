<template>
  <div class="system">
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'System'
}
</script>

<style scoped>
.system {
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>