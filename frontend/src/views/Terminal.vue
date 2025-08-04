<template>
  <div class="terminal">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2 terminal-icon">mdi-console</v-icon>
            Web Terminal
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
                Connected to Local Terminal
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
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export default {
  name: 'Terminal',
  setup() {
    const isConnected = ref(false)
    const terminalContainer = ref(null)
    let terminal = null
    let fitAddon = null
    let websocket = null
    
    // 创建终端实例
    const createTerminal = () => {
      terminal = new Terminal({
        cursorBlink: true,
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
          cursor: '#ffffff'
        },
        fontSize: 14,
        fontFamily: 'Monaco, Consolas, "Courier New", monospace'
      })
      
      fitAddon = new FitAddon()
      terminal.loadAddon(fitAddon)
      
      // 打开终端
      terminal.open(terminalContainer.value)
      fitAddon.fit()
      
      // 监听终端大小变化
      window.addEventListener('resize', () => {
        if (isConnected.value && fitAddon) {
          fitAddon.fit()
          sendResizeMessage()
        }
      })
    }
    
    // 连接到WebSocket
    const connect = () => {
      if (isConnected.value) return
      
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/ws`
      
      try {
        websocket = new WebSocket(wsUrl)
        
        websocket.onopen = () => {
          isConnected.value = true
          terminal.write('\x1b[32mConnected to terminal\x1b[0m\r\n')
        }
        
        websocket.onmessage = (event) => {
          const message = JSON.parse(event.data)
          switch (message.type) {
            case 'output':
              terminal.write(message.data)
              break
            case 'error':
              terminal.write(`\x1b[31m${message.data}\x1b[0m\r\n`)
              break
            case 'pong':
              // 心跳响应，不做处理
              break
          }
        }
        
        websocket.onclose = () => {
          isConnected.value = false
          terminal.write('\x1b[31m\n\rConnection closed\x1b[0m\r\n')
        }
        
        websocket.onerror = (error) => {
          terminal.write(`\x1b[31mConnection error: ${error.message || 'Unknown error'}\x1b[0m\r\n`)
        }
      } catch (error) {
        terminal.write(`\x1b[31mFailed to connect: ${error.message || 'Unknown error'}\x1b[0m\r\n`)
      }
    }
    
    // 断开连接
    const disconnect = () => {
      if (websocket) {
        websocket.close()
        websocket = null
      }
      isConnected.value = false
    }
    
    // 发送输入到终端
    const sendInput = (data) => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({
          type: 'input',
          data: data
        })
        websocket.send(message)
      }
    }
    
    // 发送窗口大小调整消息
    const sendResizeMessage = () => {
      if (websocket && websocket.readyState === WebSocket.OPEN && terminal && fitAddon) {
        const dims = fitAddon.proposeDimensions()
        if (dims && dims.cols > 0 && dims.rows > 0) {
          const message = JSON.stringify({
            type: 'resize',
            data: {
              cols: dims.cols,
              rows: dims.rows
            }
          })
          websocket.send(message)
        }
      }
    }
    
    onMounted(() => {
      createTerminal()
      
      // 监听终端输入
      terminal.onData((data) => {
        sendInput(data)
      })
      
      // 自动连接
      connect()
    })
    
    onBeforeUnmount(() => {
      disconnect()
      if (terminal) {
        terminal.dispose()
      }
    })
    
    return {
      isConnected,
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

.terminal-container {
  height: 500px;
  padding: 12px;
}

.terminal-container.connected {
  background-color: #000000;
}
</style>