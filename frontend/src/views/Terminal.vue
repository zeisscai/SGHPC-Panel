<template>
  <div class="terminal">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2 terminal-icon">mdi-console</v-icon>
            SSH Terminal
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2" transition="slide-y-reverse-transition">
          <v-card-text class="pa-0">
            <div class="terminal-toolbar">
              <v-btn 
                @click="connect" 
                :disabled="isConnected"
                color="success"
                small
                class="mr-2"
              >
                {{ isConnected ? 'Connected' : 'Connect' }}
              </v-btn>
              <v-btn 
                @click="disconnect" 
                :disabled="!isConnected"
                color="error"
                small
              >
                Disconnect
              </v-btn>
              <v-spacer></v-spacer>
              <v-chip v-if="isConnected" color="success" small>
                <v-icon left>mdi-check-circle</v-icon>
                Connected to {{ host }}
              </v-chip>
            </div>
            
            <div 
              ref="terminalContainer" 
              class="terminal-container"
              :class="{ 'connected': isConnected }"
            ></div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    
  </div>
</template>

<script>
import { ref, onMounted, onBeforeUnmount } from 'vue'

export default {
  name: 'Terminal',
  setup() {
    const isConnected = ref(false)
    const host = ref('localhost')
    const terminalContainer = ref(null)
    
    // 模拟终端连接
    const connect = () => {
      isConnected.value = true
      // 在实际实现中，这里会初始化 WebSocket 连接和终端实例
    }
    
    const disconnect = () => {
      isConnected.value = false
      // 在实际实现中，这里会断开 WebSocket 连接并清理终端实例
    }
    
    onMounted(() => {
      // 在实际实现中，这里会初始化终端容器
      // 例如使用 xterm.js 创建终端实例
    })
    
    onBeforeUnmount(() => {
      if (isConnected.value) {
        disconnect()
      }
    })
    
    return {
      isConnected,
      host,
      terminalContainer,
      connect,
      disconnect
    }
  }
}
</script>

<style scoped>
.terminal {
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.v-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.v-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.2) !important;
}

.terminal-toolbar {
  display: flex;
  padding: 12px;
  background-color: #f5f5f5;
  border-bottom: 1px solid #ddd;
  align-items: center;
}


.terminal-output {
  background-color: #1e1e1e;
  color: #ffffff;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.4;
  height: 500px;
  overflow-y: auto;
  padding: 16px;
  outline: none;
  white-space: pre-wrap;
  word-break: break-word;
}

.terminal-line {
  margin: 2px 0;
}

.terminal-prompt {
  color: #4ec9b0;
  margin-right: 5px;
}

.command-text {
  color: #d4d4d4;
}

.terminal-cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background-color: #ffffff;
  vertical-align: middle;
  margin-left: 2px;
}

.terminal-cursor.blink {
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.terminal-info {
  background-color: #e3f2fd;
  border-left: 4px solid #2196f3;
  padding: 16px;
  border-radius: 4px;
}

.terminal-info p {
  margin: 0;
  line-height: 1.5;
}
</style>