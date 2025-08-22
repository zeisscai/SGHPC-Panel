<template>
  <div class="terminal">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2">mdi-console</v-icon>
            Web Terminal
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2">
          <v-card-text class="pa-0">
            <div class="terminal-toolbar">
              <div class="d-flex align-center">
                <v-btn 
                  @click="connect" 
                  :disabled="isConnected || isConnecting"
                  :loading="isConnecting"
                  color="success"
                  size="small"
                  class="mr-2"
                >
                  <v-icon left>mdi-power</v-icon>
                  {{ isConnected ? '已连接' : '连接' }}
                </v-btn>
                
                <v-btn 
                  @click="disconnect" 
                  :disabled="!isConnected"
                  color="error"
                  size="small"
                  class="mr-2"
                >
                  <v-icon left>mdi-power-off</v-icon>
                  断开
                </v-btn>
                
                <v-btn 
                  @click="clearTerminal" 
                  color="warning"
                  size="small"
                  class="mr-2"
                >
                  <v-icon left>mdi-delete-sweep</v-icon>
                  清屏
                </v-btn>
              </div>
              
              <v-spacer></v-spacer>
              
              <div class="d-flex align-center">
                <v-chip 
                  v-if="isConnected" 
                  color="success" 
                  size="small"
                  variant="elevated"
                >
                  <v-icon start>mdi-check-circle</v-icon>
                  已连接
                </v-chip>
                <v-chip 
                  v-else-if="isConnecting" 
                  color="warning" 
                  size="small"
                  variant="elevated"
                >
                  <v-icon start>mdi-progress-clock</v-icon>
                  连接中...
                </v-chip>
                <v-chip 
                  v-else 
                  color="error" 
                  size="small"
                  variant="elevated"
                >
                  <v-icon start>mdi-close-circle</v-icon>
                  未连接
                </v-chip>
              </div>
            </div>
            
            <div 
              ref="terminalContainer" 
              class="terminal-container"
            ></div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export default {
  name: 'Terminal',
  setup() {
    const isConnected = ref(false)
    const isConnecting = ref(false)
    const terminalContainer = ref(null)

    let terminal = null
    let fitAddon = null
    let websocket = null
    
    // 创建终端实例
    const createTerminal = () => {
      try {
        terminal = new Terminal({
          cursorBlink: true,
          theme: {
            background: '#1e1e1e',
            foreground: '#ffffff'
          },
          fontSize: 14,
          fontFamily: 'Monaco, Consolas, "Courier New", monospace',
          scrollback: 1000,
          cursorStyle: 'block',
          convertEol: true
        })
        
        fitAddon = new FitAddon()
        terminal.loadAddon(fitAddon)
        
        if (terminalContainer.value) {
          terminal.open(terminalContainer.value)
          fitAddon.fit()
          
          // 监听终端输入
          terminal.onData((data) => {
            if (websocket && websocket.readyState === WebSocket.OPEN) {
              websocket.send(JSON.stringify({
                type: 'input',
                data: data
              }))
            }
          })
          
          terminal.write('终端已初始化，点击连接按钮开始使用\r\n')
        }
      } catch (error) {
        console.error('Failed to create terminal:', error)
      }
    }
    
    // 连接到WebSocket
    const connect = () => {
      if (isConnected.value || isConnecting.value) return
      
      isConnecting.value = true
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/ws`
      
      try {
        websocket = new WebSocket(wsUrl)
        
        websocket.onopen = () => {
          isConnected.value = true
          isConnecting.value = false
          if (terminal) {
            terminal.write('\x1b[32m连接成功!\x1b[0m\r\n')
          }
        }
        
        websocket.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            if (message.type === 'output' && terminal) {
              terminal.write(message.data)
            }
          } catch (error) {
            console.error('Failed to parse message:', error)
          }
        }
        
        websocket.onerror = (error) => {
          console.error('WebSocket error:', error)
          isConnecting.value = false
          if (terminal) {
            terminal.write('\x1b[31m连接错误\x1b[0m\r\n')
          }
        }
        
        websocket.onclose = () => {
          isConnected.value = false
          isConnecting.value = false
          if (terminal) {
            terminal.write('\x1b[33m连接已断开\x1b[0m\r\n')
          }
        }
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error)
        isConnecting.value = false
      }
    }
    
    // 断开连接
    const disconnect = () => {
      if (websocket) {
        websocket.close()
        websocket = null
      }
      isConnected.value = false
      isConnecting.value = false
    }
    
    // 清空终端
    const clearTerminal = () => {
      if (terminal) {
        terminal.clear()
      }
    }
    
    // 组件挂载时初始化
    onMounted(() => {
      setTimeout(() => {
        createTerminal()
      }, 100)
    })
    
    // 组件卸载前清理
    onBeforeUnmount(() => {
      disconnect()
    })
    
    return {
      isConnected,
      isConnecting,
      terminalContainer,
      connect,
      disconnect,
      clearTerminal
    }
  }
}
</script>

<style scoped>
.terminal {
  height: 100%;
}

.terminal-toolbar {
  display: flex;
  padding: 12px;
  border-bottom: 1px solid #ddd;
  align-items: center;
  justify-content: space-between;
}

.terminal-container {
  height: 60vh;
  min-height: 400px;
  padding: 12px;
  background-color: #1e1e1e;
}

@media (max-width: 768px) {
  .terminal-container {
    height: 50vh;
  }
  
  .terminal-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .terminal-toolbar > div {
    margin-bottom: 8px;
  }
}

.v-icon {
  background-color: transparent !important;
}
</style>