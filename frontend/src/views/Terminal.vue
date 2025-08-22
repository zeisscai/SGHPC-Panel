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
                
                <v-btn 
                  @click="copySelection" 
                  color="info"
                  size="small"
                  class="mr-2"
                >
                  <v-icon left>mdi-content-copy</v-icon>
                  复制
                </v-btn>
                
                <v-btn 
                  @click="pasteToTerminal" 
                  color="secondary"
                  size="small"
                >
                  <v-icon left>mdi-content-paste</v-icon>
                  粘贴
                </v-btn>
              </div>
              
              <v-spacer></v-spacer>
              
              <div class="d-flex align-center">
                
                
                <v-chip 
                  v-if="isConnected" 
                  color="success" 
                  size="small"
                  variant="elevated"
                  class="status-chip"
                >
                  <v-icon start>mdi-check-circle</v-icon>
                  已连接
                </v-chip>
                <v-chip 
                  v-else-if="isConnecting" 
                  color="warning" 
                  size="small"
                  variant="elevated"
                  class="status-chip"
                >
                  <v-icon start>mdi-progress-clock</v-icon>
                  连接中...
                </v-chip>
                <v-chip 
                  v-else 
                  color="error" 
                  size="small"
                  variant="elevated"
                  class="status-chip"
                >
                  <v-icon start>mdi-close-circle</v-icon>
                  未连接
                </v-chip>
              </div>
            </div>
            
            <div 
              ref="terminalContainer" 
              class="terminal-container"
              :class="{ 'connected': isConnected }"
            ></div>
            
            <div class="terminal-footer">
              <div class="d-flex align-center justify-space-between px-4 py-2">
                <div class="terminal-shortcuts">
                  <v-tooltip location="top">
                    <template v-slot:activator="{ props }">
                      <span v-bind="props" class="shortcut-item"><kbd>Ctrl</kbd> + <kbd>C</kbd> 中断</span>
                    </template>
                    中断当前命令
                  </v-tooltip>
                  
                  <v-tooltip location="top">
                    <template v-slot:activator="{ props }">
                      <span v-bind="props" class="shortcut-item"><kbd>Ctrl</kbd> + <kbd>D</kbd> 退出</span>
                    </template>
                    退出当前会话
                  </v-tooltip>
                  
                  <v-tooltip location="top">
                    <template v-slot:activator="{ props }">
                      <span v-bind="props" class="shortcut-item"><kbd>↑</kbd>/<kbd>↓</kbd> 历史</span>
                    </template>
                    浏览命令历史
                  </v-tooltip>
                </div>
                
                <div>
                  <v-btn 
                    icon="mdi-fullscreen" 
                    size="small" 
                    variant="text" 
                    @click="toggleFullscreen"
                    class="mr-2"
                  ></v-btn>
                  <v-btn 
                    icon="mdi-cog" 
                    size="small" 
                    variant="text"
                    @click="showSettings = true"
                  ></v-btn>
                </div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <!-- 设置对话框 -->
    <v-dialog v-model="showSettings" max-width="500">
      <v-card>
        <v-card-title class="text-h5">终端设置</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-select
                v-model="terminalTheme"
                label="主题"
                :items="['暗色', '亮色']"
                @change="applyTheme"
              ></v-select>
            </v-col>
            <v-col cols="12">
              <v-slider
                v-model="fontSize"
                label="字体大小"
                min="10"
                max="20"
                thumb-label
                @change="applyFontSize"
              ></v-slider>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" text @click="showSettings = false">关闭</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export default {
  name: 'Terminal',
  setup() {
    const isConnected = ref(false)
    const isConnecting = ref(false)
    const reconnectAttempts = ref(0)
    const maxReconnectAttempts = 5
    const reconnectInterval = 3000
    const terminalContainer = ref(null)
    const searchQuery = ref('')
    const showSettings = ref(false)
    const terminalTheme = ref('暗色')
    const fontSize = ref(14)
    const isFullscreen = ref(false)
    let terminal = null
    let fitAddon = null
    let websocket = null
    let pingInterval = null
    let reconnectTimeout = null
    
    // 创建终端实例
    const createTerminal = () => {
      const darkTheme = {
        background: '#1e1e1e',
        foreground: '#ffffff',
        cursor: '#ffffff',
        selection: 'rgba(255, 255, 255, 0.3)',
        black: '#000000',
        red: '#ff5555',
        green: '#50fa7b',
        yellow: '#f1fa8c',
        blue: '#bd93f9',
        magenta: '#ff79c6',
        cyan: '#8be9fd',
        white: '#bbbbbb',
        brightBlack: '#555555',
        brightRed: '#ff5555',
        brightGreen: '#50fa7b',
        brightYellow: '#f1fa8c',
        brightBlue: '#bd93f9',
        brightMagenta: '#ff79c6',
        brightCyan: '#8be9fd',
        brightWhite: '#ffffff'
      }
      
      try {
        terminal = new Terminal({
          cursorBlink: true,
          theme: darkTheme,
          fontSize: fontSize.value,
          fontFamily: 'Monaco, Consolas, "Courier New", monospace',
          scrollback: 5000,
          allowTransparency: true,
          cursorStyle: 'block',
          convertEol: true,
          rendererType: 'canvas'
        })
        
        // 添加插件
        fitAddon = new FitAddon()
        
        
        // 创建并添加WebLinksAddon
        const webLinksAddon = new WebLinksAddon()

        terminal.loadAddon(fitAddon)
        terminal.loadAddon(webLinksAddon)

        // 检查容器元素是否存在
        if (!terminalContainer.value) {
          console.error('Terminal container not found')
          return
        }
        
        // 打开终端
        terminal.open(terminalContainer.value)
        fitAddon.fit()
        
        // 写入初始消息
        terminal.write('正在初始化终端...\r\n')
        
        // 监听终端大小变化
        const resizeObserver = new ResizeObserver(() => {
          if (isConnected.value && fitAddon) {
            fitAddon.fit()
            sendResizeMessage()
          }
        })
        
        resizeObserver.observe(terminalContainer.value)
        
        // 监听窗口大小变化
        window.addEventListener('resize', () => {
          if (isConnected.value && fitAddon) {
            fitAddon.fit()
            sendResizeMessage()
          }
        })
      } catch (error) {
        console.error('Failed to create terminal:', error)
        if (terminalContainer.value) {
          terminalContainer.value.innerHTML = `<div style="color: red; padding: 10px;">
            终端初始化失败: ${error.message}
          </div>`
        }
      }
    }
    
    // 开始心跳检测
    const startPingInterval = () => {
      stopPingInterval() // 确保不会创建多个心跳
      pingInterval = setInterval(() => {
        if (websocket && websocket.readyState === WebSocket.OPEN) {
          const pingMessage = JSON.stringify({
            type: 'ping',
            data: Date.now()
          })
          websocket.send(pingMessage)
        }
      }, 30000) // 每30秒发送一次心跳
    }
    
    // 停止心跳检测
    const stopPingInterval = () => {
      if (pingInterval) {
        clearInterval(pingInterval)
        pingInterval = null
      }
    }
    
    // 清除重连定时器
    const clearReconnectTimeout = () => {
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout)
        reconnectTimeout = null
      }
    }
    
    // 尝试重新连接
    const tryReconnect = () => {
      if (reconnectAttempts.value >= maxReconnectAttempts || isConnected.value) {
        return
      }
      
      clearReconnectTimeout()
      reconnectTimeout = setTimeout(() => {
        reconnectAttempts.value++
        if (terminal) {
          terminal.write(`\x1b[33m\r\n尝试重新连接 (${reconnectAttempts.value}/${maxReconnectAttempts})...\x1b[0m\r\n`)
        }
        connect()
      }, reconnectInterval)
    }
    
    // 发送窗口大小调整消息
    const sendResizeMessage = () => {
      if (isConnected.value && websocket && websocket.readyState === WebSocket.OPEN && fitAddon) {
        const dims = fitAddon.proposeDimensions()
        if (dims && dims.cols > 0 && dims.rows > 0) {
          const resizeMessage = JSON.stringify({
            type: 'resize',
            data: {
              cols: dims.cols,
              rows: dims.rows
            }
          })
          websocket.send(resizeMessage)
        }
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
          reconnectAttempts.value = 0
          clearReconnectTimeout()
          startPingInterval()
          sendResizeMessage() // 连接成功后立即发送终端大小
          if (terminal) {
            terminal.write('\x1b[32m\r\n连接成功!\x1b[0m\r\n')
          }
        }
        
        websocket.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            switch (message.type) {
              case 'output':
                if (terminal) {
                  terminal.write(message.data)
                }
                break
              case 'error':
                if (terminal) {
                  terminal.write(`\x1b[31m${message.data}\x1b[0m\r\n`)
                }
                break
              case 'pong':
                // 心跳响应，不做处理
                break
            }
          } catch (error) {
            console.error('Failed to parse message:', error)
            if (terminal) {
              terminal.write(`\x1b[31m消息解析错误: ${error.message || '未知错误'}\x1b[0m\r\n`)
            }
          }
        }
        
        websocket.onerror = (error) => {
          console.error('WebSocket error:', error)
          isConnecting.value = false
          if (terminal) {
            terminal.write(`\x1b[31m\r\nWebSocket连接错误: ${error.message || '未知错误'}\x1b[0m\r\n`)
          }
        }
        
        websocket.onclose = (event) => {
          isConnected.value = false
          isConnecting.value = false
          stopPingInterval()
          
          if (terminal) {
            if (event.wasClean) {
              terminal.write(`\x1b[33m\r\n连接已关闭: 代码=${event.code} 原因=${event.reason || '未知'}\x1b[0m\r\n`)
            } else {
              terminal.write('\x1b[31m\r\n连接意外断开\x1b[0m\r\n')
              tryReconnect()
            }
          } else {
            console.log('WebSocket closed:', event)
            tryReconnect()
          }
        }
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error)
        isConnecting.value = false
        if (terminal) {
          terminal.write(`\x1b[31m\r\n连接失败: ${error.message || '未知错误'}\x1b[0m\r\n`)
        }
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
    
    // 复制选中内容
    const copySelection = () => {
      if (terminal) {
        const selection = terminal.getSelection()
        if (selection) {
          navigator.clipboard.writeText(selection)
        }
      }
    }
    
    // 粘贴到终端
    const pasteToTerminal = async () => {
      try {
        const text = await navigator.clipboard.readText()
        if (text && websocket && websocket.readyState === WebSocket.OPEN) {
          const inputMessage = JSON.stringify({
            type: 'input',
            data: text
          })
          websocket.send(inputMessage)
        }
      } catch (error) {
        console.error('Failed to paste text:', error)
      }
    }
    
    // 应用主题
    const applyTheme = () => {
      if (terminal) {
        const theme = terminalTheme.value === '暗色' ? {
          background: '#1e1e1e',
          foreground: '#ffffff'
        } : {
          background: '#ffffff',
          foreground: '#000000'
        }
        terminal.setOption('theme', theme)
      }
    }
    
    // 应用字体大小
    const applyFontSize = () => {
      if (terminal) {
        terminal.setOption('fontSize', fontSize.value)
        if (fitAddon) {
          fitAddon.fit()
        }
      }
    }
    
    // 切换全屏
    const toggleFullscreen = () => {
      if (!document.fullscreenElement) {
        terminalContainer.value?.requestFullscreen()
      } else {
        document.exitFullscreen()
      }
    }
    
    // 组件挂载时初始化
    onMounted(() => {
      // 确保DOM已更新后再创建终端
      setTimeout(() => {
        createTerminal()
        // 自动连接
        // connect()
      }, 100)
    })
    
    // 组件激活时（路由切换回来时）
    onActivated(() => {
      console.log('Terminal component activated')
      // 重新调整终端大小
      if (fitAddon && isConnected.value) {
        setTimeout(() => {
          fitAddon.fit()
          sendResizeMessage()
        }, 100)
      }
    })
    
    // 组件停用时
    onDeactivated(() => {
      console.log('Terminal component deactivated')
    })
    
    // 组件卸载前清理
    onBeforeUnmount(() => {
      disconnect()
      stopPingInterval()
      clearReconnectTimeout()
    })
    
    // 监听终端输入
    const handleTerminalInput = (data) => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        const inputMessage = JSON.stringify({
          type: 'input',
          data: data
        })
        websocket.send(inputMessage)
      }
    }
    
    // 监听终端大小变化
    const handleTerminalResize = (data) => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        const resizeMessage = JSON.stringify({
          type: 'resize',
          data: data
        })
        websocket.send(resizeMessage)
      }
    }
    
    return {
      isConnected,
      isConnecting,
      terminalContainer,
      showSettings,
      terminalTheme,
      fontSize,
      isFullscreen,
      connect,
      disconnect,
      clearTerminal,
      copySelection,
      pasteToTerminal,
      applyTheme,
      applyFontSize,
      toggleFullscreen,
      handleTerminalInput,
      handleTerminalResize
    }
  }
}
</script>

<style scoped>
.terminal {
  animation: fadeIn 0.5s ease-in;
  height: 100%;
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
  justify-content: space-between;
  flex-wrap: wrap;
}

.terminal-container {
  height: 65vh; /* 使用视口高度 */
  min-height: 400px; /* 最小高度 */
  padding: 12px;
  overflow: hidden;
  background-color: #1e1e1e; /* 添加默认背景色 */
}

.terminal-container.connected {
  background-color: #000000;
}

.terminal-footer {
  background-color: #f5f5f5;
  border-top: 1px solid #ddd;
}

.shortcut-item {
  margin-right: 12px;
  font-size: 0.8rem;
  color: #666;
}

kbd {
  background-color: #eee;
  border-radius: 3px;
  border: 1px solid #b4b4b4;
  box-shadow: 0 1px 1px rgba(0, 0, 0, .2);
  color: #333;
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
  padding: 2px 4px;
  white-space: nowrap;
}

.search-field {
  max-width: 200px;
}

.status-chip {
  min-width: 80px;
  justify-content: center;
}

/* 全屏模式样式 */
.terminal-container:fullscreen {
  padding: 0;
  width: 100vw;
  height: 100vh;
  background-color: #000;
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
  
  .shortcut-item {
    display: none;
  }
}
</style>