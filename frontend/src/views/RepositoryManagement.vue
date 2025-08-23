<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card elevation="2">
          <v-card-title>
            <v-icon left style="background: transparent;">mdi-package-variant</v-icon>
            Repository
          </v-card-title>
          
          <v-card-text>


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
                  color="info"
                  block
                  @click="showEditDialog = true"
                >
                  <v-icon left style="background: transparent;">mdi-pencil</v-icon>
                  自定义编辑
                </v-btn>
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
  setCustomRepository
} from '../api/repository'

export default {
  name: 'RepositoryManagement',
  setup() {
    // 加载状态
    const cleaningCache = ref(false)
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
    
    // 显示成功消息
    const showSuccess = (message) => {
      successMessage.value = message
      showSuccessMessage.value = true
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
      
      try {
        const response = await cleanRepositoryCache()
        if (response.success) {
          showSuccess('缓存清理完成')
        }
      } catch (error) {
        // 静默处理错误
      } finally {
        cleaningCache.value = false
      }
    }
    
    
    // 应用自定义软件源
    const applyCustomRepository = async () => {
      if (!isValidUrl(customRepoUrl.value)) {
        return
      }
      
      applyingCustomRepo.value = true
      
      try {
        const response = await setCustomRepository(customRepoUrl.value)
        if (response.success) {
          showSuccess('自定义软件源应用成功')
          showEditDialog.value = false
          customRepoUrl.value = ''
        }
      } catch (error) {
        // 静默处理错误
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
      // 加载状态
      cleaningCache,
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