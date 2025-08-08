<template>
  <div>
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon left class="mr-2">mdi-package-variant</v-icon>
            Spack Package Manager
          </v-card-title>
          <v-card-text>
            <!-- Spack 安装状态检查 -->
            <v-alert
              v-if="spackStatus === 'not_installed'"
              type="warning"
              outlined
            >
              <strong>Spack 未安装</strong>
              <div class="mt-2">
                <v-btn
                  color="primary"
                  @click="installSpack"
                  :loading="isInstalling"
                >
                  <v-icon left>mdi-download</v-icon>
                  安装 Spack 1.0.0
                </v-btn>
              </div>
            </v-alert>
            
            <v-alert
              v-else-if="spackStatus === 'installed'"
              type="success"
              outlined
            >
              <strong>Spack 已安装</strong>
              <div class="mt-2">
                <p>版本: {{ spackVersion }}</p>
              </div>
            </v-alert>
            
            <!-- Spack 功能区域 -->
            <div v-if="spackStatus === 'installed'">
              <v-row class="mb-4">
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="searchQuery"
                    label="搜索软件包"
                    outlined
                    dense
                    clearable
                    append-icon="mdi-magnify"
                    @click:append="searchPackages"
                    @keyup.enter="searchPackages"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-btn
                    color="primary"
                    @click="refreshPackageLists"
                    :loading="isRefreshing"
                    class="mr-2"
                  >
                    <v-icon left>mdi-refresh</v-icon>
                    刷新列表
                  </v-btn>
                  <v-btn
                    color="secondary"
                    @click="modifyRepositories"
                  >
                    <v-icon left>mdi-source-repository</v-icon>
                    修改软件源
                  </v-btn>
                </v-col>
              </v-row>
              
              <v-tabs v-model="activeTab" fixed-tabs>
                <v-tab key="available">可安装</v-tab>
                <v-tab key="installed">已安装</v-tab>
              </v-tabs>
              
              <v-tabs-items v-model="activeTab">
                <!-- 可安装软件包选项卡 -->
                <v-tab-item key="available">
                  <v-card flat>
                    <v-card-text>
                      <v-data-table
                        :headers="availablePackagesHeaders"
                        :items="filteredAvailablePackages"
                        :loading="loadingAvailablePackages"
                        loading-text="加载中..."
                        no-data-text="未找到可安装的软件包"
                        :items-per-page="10"
                        class="elevation-1"
                      >
                        <template v-slot:item.actions="{ item }">
                          <v-btn
                            color="primary"
                            small
                            @click="installPackage(item, 'normal')"
                            class="mr-2"
                          >
                            安装
                          </v-btn>
                          <v-btn
                            color="secondary"
                            small
                            @click="installPackage(item, 'advanced')"
                          >
                            高级安装
                          </v-btn>
                        </template>
                      </v-data-table>
                    </v-card-text>
                  </v-card>
                </v-tab-item>
                
                <!-- 已安装软件包选项卡 -->
                <v-tab-item key="installed">
                  <v-card flat>
                    <v-card-text>
                      <v-data-table
                        :headers="installedPackagesHeaders"
                        :items="filteredInstalledPackages"
                        :loading="loadingInstalledPackages"
                        loading-text="加载中..."
                        no-data-text="未找到已安装的软件包"
                        :items-per-page="10"
                        class="elevation-1"
                      >
                        <template v-slot:item.actions="{ item }">
                          <v-btn
                            color="error"
                            small
                            @click="uninstallPackage(item)"
                          >
                            卸载
                          </v-btn>
                        </template>
                      </v-data-table>
                    </v-card-text>
                  </v-card>
                </v-tab-item>
              </v-tabs-items>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <!-- 安装日志对话框 -->
    <v-dialog v-model="showLogDialog" max-width="800" persistent>
      <v-card>
        <v-card-title>
          <span class="headline">安装日志</span>
        </v-card-title>
        <v-card-text>
          <v-textarea
            outlined
            readonly
            :rows="15"
            v-model="installLog"
            ref="logTextarea"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            @click="closeLogDialog"
            :disabled="!installCompleted"
          >
            关闭
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 高级安装对话框 -->
    <v-dialog v-model="showAdvancedInstallDialog" max-width="600" persistent>
      <v-card>
        <v-card-title>
          <span class="headline">高级安装</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="advancedInstallOptions"
            label="安装选项"
            outlined
            dense
          ></v-text-field>
          <p>示例: +mpi +python ^openmpi</p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="secondary"
            @click="showAdvancedInstallDialog = false"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            @click="confirmAdvancedInstall"
          >
            安装
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 软件源修改对话框 -->
    <v-dialog v-model="showRepoDialog" max-width="600" persistent>
      <v-card>
        <v-card-title>
          <span class="headline">修改软件源</span>
        </v-card-title>
        <v-card-text>
          <v-textarea
            v-model="repositories"
            label="软件源配置"
            outlined
            rows="10"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="secondary"
            @click="showRepoDialog = false"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            @click="saveRepositories"
          >
            保存
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'

export default {
  name: 'Spack',
  setup() {
    // Spack 状态 (not_installed, installed)
    const spackStatus = ref('checking')
    const spackVersion = ref('')
    
    // 搜索查询
    const searchQuery = ref('')
    
    // 活动选项卡
    const activeTab = ref(0)
    
    // 加载状态
    const isInstalling = ref(false)
    const isRefreshing = ref(false)
    const loadingAvailablePackages = ref(false)
    const loadingInstalledPackages = ref(false)
    
    // 软件包列表
    const availablePackages = ref([])
    const installedPackages = ref([])
    
    // 对话框状态
    const showLogDialog = ref(false)
    const showAdvancedInstallDialog = ref(false)
    const showRepoDialog = ref(false)
    
    // 安装相关
    const installLog = ref('')
    const installCompleted = ref(false)
    const currentPackage = ref(null)
    const advancedInstallOptions = ref('')
    const repositories = ref('')
    
    // WebSocket 连接
    const ws = ref(null)
    
    // 表格列定义
    const availablePackagesHeaders = [
      { text: '名称', value: 'name' },
      { text: '描述', value: 'description' },
      { text: '操作', value: 'actions', sortable: false }
    ]
    
    const installedPackagesHeaders = [
      { text: '名称', value: 'name' },
      { text: '版本', value: 'version' },
      { text: '哈希', value: 'hash' },
      { text: '操作', value: 'actions', sortable: false }
    ]
    
    // 过滤后的软件包列表
    const filteredAvailablePackages = computed(() => {
      if (!searchQuery.value) {
        return availablePackages.value
      }
      const query = searchQuery.value.toLowerCase()
      return availablePackages.value.filter(pkg => 
        pkg.name.toLowerCase().includes(query) || 
        (pkg.description && pkg.description.toLowerCase().includes(query))
      )
    })
    
    const filteredInstalledPackages = computed(() => {
      if (!searchQuery.value) {
        return installedPackages.value
      }
      const query = searchQuery.value.toLowerCase()
      return installedPackages.value.filter(pkg => 
        pkg.name.toLowerCase().includes(query)
      )
    })
    
    // 检查 Spack 安装状态
    const checkSpackStatus = async () => {
      try {
        const response = await fetch('/api/spack/status')
        const data = await response.json()
        spackStatus.value = data.installed ? 'installed' : 'not_installed'
        spackVersion.value = data.version
      } catch (error) {
        console.error('检查 Spack 状态失败:', error)
        spackStatus.value = 'not_installed'
      }
    }
    
    // 安装 Spack
    const installSpack = async () => {
      isInstalling.value = true
      showLogDialog.value = true
      installLog.value = '正在连接到安装服务...\n'
      installCompleted.value = false
      
      try {
        // 连接到 WebSocket 端点以获取实时日志
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsUrl = `${protocol}//${window.location.host}/api/spack/install/logs`
        ws.value = new WebSocket(wsUrl)
        
        ws.value.onopen = () => {
          installLog.value += '已连接到安装服务，开始安装...\n'
        }
        
        ws.value.onmessage = (event) => {
          if (event.data === 'INSTALL_COMPLETED') {
            installLog.value += '\n安装完成！\n'
            installCompleted.value = true
            isInstalling.value = false
            // 重新检查 Spack 状态
            setTimeout(checkSpackStatus, 1000)
          } else {
            installLog.value += event.data + '\n'
            // 自动滚动到底部
            setTimeout(() => {
              const textarea = document.querySelector('.v-dialog--active textarea')
              if (textarea) {
                textarea.scrollTop = textarea.scrollHeight
              }
            }, 100)
          }
        }
        
        ws.value.onerror = (error) => {
          installLog.value += `\n连接错误: ${error.message}\n`
          installCompleted.value = true
          isInstalling.value = false
        }
        
        ws.value.onclose = () => {
          // 连接关闭是正常的
        }
      } catch (error) {
        installLog.value += `安装失败: ${error.message}\n`
        installCompleted.value = true
        isInstalling.value = false
      }
    }
    
    // 刷新软件包列表
    const refreshPackageLists = async () => {
      isRefreshing.value = true
      try {
        // 获取可安装的软件包
        loadingAvailablePackages.value = true
        const availableResponse = await fetch('/api/spack/packages/available')
        if (availableResponse.ok) {
          availablePackages.value = await availableResponse.json()
        } else {
          console.error('获取可安装软件包列表失败:', await availableResponse.text())
        }
        loadingAvailablePackages.value = false
        
        // 获取已安装的软件包
        loadingInstalledPackages.value = true
        const installedResponse = await fetch('/api/spack/packages/installed')
        if (installedResponse.ok) {
          installedPackages.value = await installedResponse.json()
        } else {
          console.error('获取已安装软件包列表失败:', await installedResponse.text())
        }
        loadingInstalledPackages.value = false
      } catch (error) {
        console.error('刷新软件包列表失败:', error)
        loadingAvailablePackages.value = false
        loadingInstalledPackages.value = false
      } finally {
        isRefreshing.value = false
      }
    }
    
    // 搜索软件包
    const searchPackages = () => {
      // 搜索功能已经在 computed 属性中实现
      // 这里可以添加额外的搜索逻辑
    }
    
    // 安装软件包
    const installPackage = (packageItem, type) => {
      currentPackage.value = packageItem
      if (type === 'normal') {
        performInstall(packageItem.name)
      } else {
        advancedInstallOptions.value = ''
        showAdvancedInstallDialog.value = true
      }
    }
    
    // 执行安装
    const performInstall = async (packageName, options = '') => {
      showAdvancedInstallDialog.value = false
      showLogDialog.value = true
      installLog.value = `正在准备安装 ${packageName}...\n`
      installCompleted.value = false
      
      try {
        // 连接到 WebSocket 端点以获取实时日志
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsUrl = `${protocol}//${window.location.host}/api/spack/package/install/logs?package=${encodeURIComponent(packageName)}&options=${encodeURIComponent(options)}`
        ws.value = new WebSocket(wsUrl)
        
        ws.value.onopen = () => {
          installLog.value += `已连接到安装服务，开始安装 ${packageName}...\n`
        }
        
        ws.value.onmessage = (event) => {
          if (event.data === 'INSTALL_COMPLETED') {
            installLog.value += `\n${packageName} 安装完成！\n`
            installCompleted.value = true
            // 刷新已安装软件包列表
            setTimeout(refreshPackageLists, 1000)
          } else {
            installLog.value += event.data + '\n'
            // 自动滚动到底部
            setTimeout(() => {
              const textarea = document.querySelector('.v-dialog--active textarea')
              if (textarea) {
                textarea.scrollTop = textarea.scrollHeight
              }
            }, 100)
          }
        }
        
        ws.value.onerror = (error) => {
          installLog.value += `\n连接错误: ${error.message}\n`
          installCompleted.value = true
        }
        
        ws.value.onclose = () => {
          // 连接关闭是正常的
        }
      } catch (error) {
        installLog.value += `安装失败: ${error.message}\n`
        installCompleted.value = true
      }
    }
    
    // 确认高级安装
    const confirmAdvancedInstall = () => {
      performInstall(
        currentPackage.value.name, 
        advancedInstallOptions.value
      )
    }
    
    // 卸载软件包
    const uninstallPackage = async (packageItem) => {
      if (!confirm(`确定要卸载 ${packageItem.name} 吗?`)) {
        return
      }
      
      try {
        const response = await fetch('/api/spack/package/uninstall', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            package_name: packageItem.name
          })
        })
        
        if (response.ok) {
          // 刷新已安装软件包列表
          refreshPackageLists()
        } else {
          console.error('卸载软件包失败:', await response.text())
        }
      } catch (error) {
        console.error('卸载软件包失败:', error)
      }
    }
    
    // 修改软件源
    const modifyRepositories = async () => {
      try {
        const response = await fetch('/api/spack/repositories')
        if (response.ok) {
          const data = await response.json()
          repositories.value = data.content
          showRepoDialog.value = true
        } else {
          console.error('获取软件源配置失败:', await response.text())
        }
      } catch (error) {
        console.error('获取软件源配置失败:', error)
      }
    }
    
    // 保存软件源配置
    const saveRepositories = async () => {
      try {
        const response = await fetch('/api/spack/repositories/update', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            content: repositories.value
          })
        })
        
        if (response.ok) {
          showRepoDialog.value = false
        } else {
          console.error('保存软件源配置失败:', await response.text())
        }
      } catch (error) {
        console.error('保存软件源配置失败:', error)
      }
    }
    
    // 关闭日志对话框
    const closeLogDialog = () => {
      showLogDialog.value = false
      installLog.value = ''
      installCompleted.value = false
      
      // 关闭 WebSocket 连接
      if (ws.value) {
        ws.value.close()
        ws.value = null
      }
    }
    
    // 组件挂载时检查 Spack 状态
    onMounted(() => {
      checkSpackStatus()
    })
    
    return {
      spackStatus,
      spackVersion,
      searchQuery,
      activeTab,
      isInstalling,
      isRefreshing,
      loadingAvailablePackages,
      loadingInstalledPackages,
      availablePackages,
      installedPackages,
      showLogDialog,
      showAdvancedInstallDialog,
      showRepoDialog,
      installLog,
      installCompleted,
      currentPackage,
      advancedInstallOptions,
      repositories,
      availablePackagesHeaders,
      installedPackagesHeaders,
      filteredAvailablePackages,
      filteredInstalledPackages,
      checkSpackStatus,
      installSpack,
      refreshPackageLists,
      searchPackages,
      installPackage,
      uninstallPackage,
      modifyRepositories,
      saveRepositories,
      confirmAdvancedInstall,
      closeLogDialog
    }
  }
}
</script>

<style scoped>
.v-data-table {
  margin-top: 16px;
}
</style>