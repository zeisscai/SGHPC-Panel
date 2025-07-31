<template>
  <div class="card">
    <h2>部署日志</h2>
    <div class="log-container">
      <pre>{{ logs }}</pre>
    </div>
    <div style="margin-top: 1rem;">
      <button @click="clearLogs" class="btn">清空日志</button>
      <button @click="downloadLogs" class="btn">下载日志</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LogViewer',
  props: {
    logs: {
      type: String,
      default: ''
    }
  },
  methods: {
    clearLogs() {
      this.$emit('clear-logs');
    },
    
    downloadLogs() {
      this.$emit('download-logs');
    }
  }
}
</script>

<style scoped>
.log-container {
  background-color: #000;
  color: #00ff00;
  padding: 1rem;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>