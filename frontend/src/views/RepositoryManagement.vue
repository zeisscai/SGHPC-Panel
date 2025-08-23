<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card elevation="2">
          <v-card-title>
            <v-icon left style="background: transparent;">mdi-package-variant</v-icon>
            软件源管理
          </v-card-title>
          
          <v-card-text>
            <!-- 当前状态显示 -->
            <v-row class="mb-4">
              <v-col cols="12">
                <v-card outlined>
                  <v-card-subtitle>当前软件源状态</v-card-subtitle>
                  <v-card-text>
                    <v-row>
                      <v-col cols="12" md="6">
                        <v-chip
                          :color="currentRepoType === 'default' ? 'primary' : 'secondary'"
                          label
                          class="mb-2"
                        >
                          <v-icon left style="background: transparent;">mdi-source-repository</v-icon>
                          {{ currentRepoType === 'default' ? '默认软件源' : '中科大镜像源' }}
                        </v-chip>
                        <div class="text-caption mt-1">
                          {{ currentRepoUrl }}
                        </div>
                      </v-col>
                      <v-col cols="12" md="6">
                        <v-chip
                          :color="repoStatus === 'accessible' ? 'success' : repoStatus === 'checking' ? 'warning' : 'error'"
                          label
                        >
                          <v-icon left style="background: transparent;">
                            {{ repoStatus === 'accessible' ? 'mdi-check-circle' : repoStatus === 'checking' ? 'mdi-loading mdi-spin' : 'mdi-alert-circle' }}
                          </v-icon>
                          {{ repoStatus === 'accessible' ? '可访问' : repoStatus === 'checking' ? '检查中' : '不可访问' }}
                        </v-chip>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>

            <!-- 操作按钮区域 -->
            <v-row class="mb-4">
              <v-col cols="12" md="3">
                <v-btn
                  color="primary"
                  block
                  :loading="cleaningCache"
                  @click="cleanCache"
                >
                  <v-icon left style="background: transparent;">mdi-broom</v-icon>
                  清理缓存
                </v-btn>
              </v-col>
              <v-col cols="12" md="3">
                <v-btn
                  color="secondary"
                  block
                  :loading="switchingRepo"
                  @click="switchRepository"
                >
                  <v-icon left style="background: transparent;">mdi-swap-horizontal</v-icon>
                  切换软件源
                </v-btn>
              </v-col>
              <v-col cols="12" md="3">
                <v-btn
                  color="info"
                  block
                  @click="showEditDialog = true"
                >
                  <v-icon left style="background: transparent;">mdi-pencil</v-icon>
                  自定义编辑
                </v-btn>
              </v-col>

            </v-row>

            <!-- 操作日志 -->
            <v-row>
              <v-col cols="12">
                <v-card outlined>
                  <v-card-subtitle>操作日志</v-card-subtitle>
                  <v-card-text>
                    <v-textarea
                      v-model="operationLog"
                      readonly
                      outlined
                      rows="8"
                      placeholder="操作日志将在此显示..."
                      class="monospace-font"
                    ></v-textarea>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- 自定义编辑对话框 -->
    <v-dialog v-model="showEditDialog" max-width="600" persistent>
      <v-card>
        <v-card-title>
          <span class="headline">自定义软件源</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="customRepoUrl"
            label="软件源链接"
            placeholder="https://mirrors.ustc.edu.cn/"
            outlined
            dense
            :rules="[rules.required, rules.url]"
          ></v-text-field>
          <v-alert
            type="info"
            outlined
            dense
            class="mt-3"
          >
            <div class="text-caption">
              <strong>说明：</strong><br>
              • 请输入完整的镜像站根URL<br>
              • 系统将自动生成所有repo配置文件<br>
              • 示例：https://mirrors.ustc.edu.cn/
            </div>
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="secondary"
            @click="cancelEdit"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            :loading="applyingCustomRepo"
            @click="applyCustomRepository"
            :disabled="!isValidUrl(customRepoUrl)"
          >
            应用
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 成功提示 -->
    <v-snackbar
      v-model="showSuccessMessage"
      color="success"
      timeout="3000"
    >
      {{ successMessage }}
      <template v-slot:action="{ attrs }">
        <v-btn
          text
          v-bind="attrs"
          @click="showSuccessMessage = false"
        >
          关闭
        </v-btn>
      </template>
    </v-snackbar>


  </v-container>
</template>

<script>
import { ref, onMounted } from 'vue'
import {
  cleanRepositoryCache,
  switchRepository,
  setCustomRepository
} from '../api/repository'

export default {
  name: 'RepositoryManagement',
  setup() {
    // 响应式数据
    const currentRepoType = ref('default')
    const currentRepoUrl = ref('')
    const repoStatus = ref('checking')
    const operationLog = ref('')
    
    // 加载状态
    const cleaningCache = ref(false)
    const switchingRepo = ref(false)
    const applyingCustomRepo = ref(false)
    
    // 对话框状态
    const showEditDialog = ref(false)
    const customRepoUrl = ref('')
    
    // 消息提示
    const showSuccessMessage = ref(false)
    const successMessage = ref('')
    
    // 表单验证规则
    const rules = {
      required: value => !!value || '此字段为必填项',
      url: value => {
        try {
          new URL(value)
          return true
        } catch {
          return '请输入有效的URL'
        }
      }
    }
    
    // 添加日志
    const addLog = (message) => {
      const timestamp = new Date().toLocaleString()
      operationLog.value += `[${timestamp}] ${message}\n`
    }
    
    // 显示成功消息
    const showSuccess = (message) => {
      successMessage.value = message
      showSuccessMessage.value = true
      addLog(`✓ ${message}`)
    }
    

    
    // 检查URL有效性
    const isValidUrl = (url) => {
      try {
        new URL(url)
        return true
      } catch {
        return false
      }
    }
    
    // 清理缓存
    const cleanCache = async () => {
      cleaningCache.value = true
      addLog('开始清理软件包缓存...')
      
      try {
        const response = await cleanRepositoryCache()
        if (response.success) {
          showSuccess('缓存清理完成')
        } else {
          addLog(`✗ ${response.message || '缓存清理失败'}`)
        }
      } catch (error) {
        addLog(`✗ 缓存清理失败：${error.message}`)
      } finally {
        cleaningCache.value = false
      }
    }
    
    // 切换软件源
    const switchRepository = async () => {
      switchingRepo.value = true
      const targetType = currentRepoType.value === 'default' ? 'ustc' : 'default'
      addLog(`正在切换到${targetType === 'default' ? '默认' : '中科大镜像'}软件源...`)
      
      try {
        const response = await switchRepository(targetType)
        if (response.success) {
          showSuccess('软件源切换成功')
        } else {
          addLog(`✗ ${response.message || '软件源切换失败'}`)
        }
      } catch (error) {
        addLog(`✗ 软件源切换失败：${error.message}`)
      } finally {
        switchingRepo.value = false
      }
    }
    

    
    // 应用自定义软件源
    const applyCustomRepository = async () => {
      if (!isValidUrl(customRepoUrl.value)) {
        addLog('✗ 请输入有效的URL')
        return
      }
      
      applyingCustomRepo.value = true
      addLog(`正在应用自定义软件源：${customRepoUrl.value}`)
      
      try {
        const response = await setCustomRepository(customRepoUrl.value)
        if (response.success) {
          showSuccess('自定义软件源应用成功')
          showEditDialog.value = false
          customRepoUrl.value = ''
        } else {
          addLog(`✗ ${response.message || '自定义软件源应用失败'}`)
        }
      } catch (error) {
        addLog(`✗ 自定义软件源应用失败：${error.message}`)
      } finally {
        applyingCustomRepo.value = false
      }
    }
    
    // 取消编辑
    const cancelEdit = () => {
      showEditDialog.value = false
      customRepoUrl.value = ''
    }
    

    
    return {
      // 数据
      currentRepoType,
      currentRepoUrl,
      repoStatus,
      operationLog,
      
      // 加载状态
      cleaningCache,
      switchingRepo,
      applyingCustomRepo,
      
      // 对话框
      showEditDialog,
      customRepoUrl,
      
      // 消息
      showSuccessMessage,
      successMessage,
      
      // 验证规则
      rules,
      
      // 方法
      cleanCache,
      switchRepository,
      applyCustomRepository,
      cancelEdit,
      isValidUrl
    }
  }
}
</script>

<style scoped>
.monospace-font {
  font-family: 'Courier New', monospace;
}

.v-card {
  margin-bottom: 16px;
}

.v-chip {
  margin-right: 8px;
}

.mdi-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>