<template>
  <div class="terminal">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2">mdi-console</v-icon>
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
                color="primary"
              >
                {{ isConnected ? 'Connected' : 'Connect to Localhost' }}
              </v-btn>
              <v-btn 
                @click="disconnect" 
                :disabled="!isConnected"
                class="ml-2"
              >
                Disconnect
              </v-btn>
              <v-chip 
                :color="isConnected ? 'success' : 'error'" 
                class="ml-4"
                small
              >
                {{ isConnected ? 'Connected' : 'Disconnected' }}
              </v-chip>
            </div>
            
            <div 
              ref="terminalOutput" 
              class="terminal-output"
              tabindex="0"
              @keydown="handleKeyDown"
            >
              <div 
                v-for="(line, index) in terminalLines" 
                :key="index" 
                class="terminal-line"
              >
                <span v-html="line"></span>
              </div>
              <div class="terminal-line" v-if="isConnected">
                <span class="terminal-prompt">{{ prompt }}</span>
                <span class="command-text">{{ currentCommand }}</span>
                <span class="terminal-cursor" :class="{ 'blink': showCursor }"></span>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    
  </div>
</template>

<script>
export default {
  name: 'Terminal',
  data() {
    return {
      isConnected: false,
      websocket: null,
      terminalLines: [],
      currentCommand: '',
      prompt: '[user@localhost ~]$ ',
      showCursor: true,
      commandHistory: [],
      historyIndex: -1
    }
  },
  mounted() {
    this.startCursorBlink();
    // 组件加载后自动连接
    this.connect();
  },
  beforeUnmount() {
    this.disconnect();
  },
  methods: {
    connect() {
      if (this.isConnected) return;
      
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/ws`;
      
      try {
        this.websocket = new WebSocket(wsUrl);
        
        this.websocket.onopen = () => {
          this.isConnected = true;
          this.terminalLines.push('Connected to local SSH server');
          this.focusTerminal();
        };
        
        this.websocket.onmessage = (event) => {
          // 处理接收到的消息
          const data = event.data;
          if (typeof data === 'string') {
            this.terminalLines.push(data);
          } else {
            // 如果是二进制数据，转换为文本
            const reader = new FileReader();
            reader.onload = (e) => {
              this.terminalLines.push(e.target.result);
            };
            reader.readAsText(data);
          }
          this.focusTerminal();
        };
        
        this.websocket.onclose = () => {
          this.isConnected = false;
          this.terminalLines.push('Connection closed');
        };
        
        this.websocket.onerror = (error) => {
          this.terminalLines.push(`Connection error: ${error.message || error}`);
        };
      } catch (error) {
        this.terminalLines.push(`Failed to connect: ${error.message || 'Unknown error'}`);
      }
    },
    
    disconnect() {
      if (this.websocket) {
        this.websocket.close();
        this.websocket = null;
      }
      this.isConnected = false;
    },
    
    focusTerminal() {
      this.$nextTick(() => {
        if (this.$refs.terminalOutput) {
          this.$refs.terminalOutput.focus();
          // 滚动到底部
          this.$refs.terminalOutput.scrollTop = this.$refs.terminalOutput.scrollHeight;
        }
      });
    },
    
    startCursorBlink() {
      setInterval(() => {
        this.showCursor = !this.showCursor;
      }, 500);
    },
    
    handleKeyDown(event) {
      if (!this.isConnected) return;
      
      // 阻止默认行为
      event.preventDefault();
      
      // 处理特殊键
      switch (event.key) {
        case 'Enter':
          this.sendCommand();
          break;
        case 'Backspace':
          this.handleBackspace();
          break;
        case 'ArrowUp':
          this.handleArrowUp();
          break;
        case 'ArrowDown':
          this.handleArrowDown();
          break;
        default:
          // 处理普通字符输入
          if (event.key.length === 1) {
            this.currentCommand += event.key;
            this.sendToTerminal(event.key);
          }
          break;
      }
      
      // 保持焦点并滚动到底部
      this.focusTerminal();
    },
    
    sendToTerminal(data) {
      if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
        this.websocket.send(data);
      }
    },
    
    sendCommand() {
      if (this.currentCommand.trim() !== '') {
        // 添加到历史记录
        this.commandHistory.push(this.currentCommand);
        this.historyIndex = -1;
      }
      
      // 发送回车符
      this.sendToTerminal('\n');
      this.currentCommand = '';
    },
    
    handleBackspace() {
      if (this.currentCommand.length > 0) {
        this.currentCommand = this.currentCommand.slice(0, -1);
        this.sendToTerminal('\b');
      }
    },
    
    handleArrowUp() {
      if (this.commandHistory.length > 0) {
        if (this.historyIndex === -1) {
          // 保存当前输入
          this.historyIndex = this.commandHistory.length - 1;
        } else if (this.historyIndex > 0) {
          this.historyIndex--;
        }
        this.currentCommand = this.commandHistory[this.historyIndex] || '';
      }
    },
    
    handleArrowDown() {
      if (this.commandHistory.length > 0 && this.historyIndex !== -1) {
        if (this.historyIndex < this.commandHistory.length - 1) {
          this.historyIndex++;
          this.currentCommand = this.commandHistory[this.historyIndex];
        } else {
          this.historyIndex = -1;
          this.currentCommand = '';
        }
      }
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