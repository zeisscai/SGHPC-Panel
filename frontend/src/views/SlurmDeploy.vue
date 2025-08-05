<template>
  <div>
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon left class="mr-2">mdi-server-plus</v-icon>
            Slurm Deployment
          </v-card-title>
          <v-card-text>
            <v-tabs v-model="activeTab" fixed-tabs>
              <v-tab key="status">Slurm Status</v-tab>
              <v-tab key="clientStatus">Slurm Client Status</v-tab>
              <v-tab key="log">Deployment Log</v-tab>
            </v-tabs>

            <v-tabs-items v-model="activeTab">
              <!-- Slurm Status Tab -->
              <v-tab-item key="status">
                <v-card flat>
                  <v-card-text>
                    <v-alert
                      v-if="slurmStatus === 'not_installed'"
                      type="info"
                      outlined
                      colored-border
                      border="left"
                    >
                      <strong>Slurmctld:</strong> 未安装
                    </v-alert>
                    
                    <v-alert
                      v-else-if="slurmStatus === 'running'"
                      type="success"
                      outlined
                      colored-border
                      border="left"
                    >
                      <strong>Slurmctld:</strong> 正在运行
                    </v-alert>
                    
                    <v-alert
                      v-else-if="slurmStatus === 'stopped'"
                      type="error"
                      outlined
                      colored-border
                      border="left"
                    >
                      <strong>Slurmctld:</strong> 未运行
                    </v-alert>
                    
                    <v-row class="mt-4">
                      <v-col cols="12">
                        <div v-if="slurmStatus === 'not_installed'">
                          <v-btn
                            color="primary"
                            @click="startInstallation"
                            :loading="isInstalling"
                          >
                            <v-icon left>mdi-download</v-icon>
                            开始安装
                          </v-btn>
                        </div>
                        
                        <div v-else>
                          <v-btn
                            color="warning"
                            @click="controlService('restart')"
                            class="mr-2"
                          >
                            <v-icon left>mdi-restart</v-icon>
                            重启
                          </v-btn>
                          <v-btn
                            color="error"
                            @click="controlService('stop')"
                          >
                            <v-icon left>mdi-stop</v-icon>
                            停止
                          </v-btn>
                        </div>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-tab-item>

              <!-- Slurm Client Status Tab -->
              <v-tab-item key="clientStatus">
                <v-card flat>
                  <v-card-text>
                    <v-alert type="info" outlined>
                      <p>Slurm客户端状态信息将在此显示</p>
                    </v-alert>
                  </v-card-text>
                </v-card>
              </v-tab-item>

              <!-- Log Tab -->
              <v-tab-item key="log">
                <v-card flat>
                  <v-card-text>
                    <v-textarea
                      outlined
                      :model-value="formatLogs(deploymentLogs)"
                      rows="15"
                      readonly
                      no-resize
                      ref="logContainer"
                    ></v-textarea>
                    
                    <v-btn
                      color="secondary"
                      @click="refreshLogs"
                      class="mt-2"
                    >
                      <v-icon left>mdi-refresh</v-icon>
                      Refresh Logs
                    </v-btn>
                  </v-card-text>
                </v-card>
              </v-tab-item>
            </v-tabs-items>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon left class="mr-2">mdi-cog</v-icon>
            Slurm Configuration
          </v-card-title>
          <v-card-text>
            <v-form ref="configForm">
              <v-row>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="slurmConfig.cluster_name"
                    label="Cluster Name"
                    outlined
                    dense
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="slurmConfig.control_machine"
                    label="Control Machine"
                    outlined
                    dense
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="slurmConfig.control_addr"
                    label="Control Address"
                    outlined
                    dense
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="slurmConfig.slurm_user"
                    label="Slurm User"
                    outlined
                    dense
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model.number="slurmConfig.slurmctld_port"
                    label="Slurmctld Port"
                    outlined
                    dense
                    type="number"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model.number="slurmConfig.slurmd_port"
                    label="Slurmd Port"
                    outlined
                    dense
                    type="number"
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-text-field
                    v-model="slurmConfig.compute_nodes"
                    label="Compute Nodes (comma separated)"
                    outlined
                    dense
                  ></v-text-field>
                </v-col>
              </v-row>
              
              <v-btn
                color="primary"
                @click="updateConfiguration"
              >
                <v-icon left>mdi-content-save</v-icon>
                Save Configuration
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { 
  fetchSlurmConfig, 
  updateSlurmConfig, 
  fetchDeploymentStatus, 
  startDeployment,
  fetchDeploymentLogs,
  controlSlurmService,
  fetchSlurmStatus
} from '../api/slurm'

export default {
  name: 'SlurmDeploy',
  setup() {
    // Slurm配置
    const slurmConfig = ref({
      cluster_name: '',
      control_machine: '',
      control_addr: '',
      slurm_user: '',
      slurmctld_port: 6817,
      slurmd_port: 6818,
      compute_nodes: []
    })
    
    // 活动选项卡
    const activeTab = ref(0)
    
    // Slurm状态 (not_installed, running, stopped)
    const slurmStatus = ref('not_installed')
    
    // 部署状态
    const deploymentStatus = ref({
      phase: 'idle',
      progress: 0,
      message: '',
      error_message: '',
      start_time: '',
      end_time: ''
    })
    
    // 部署日志
    const deploymentLogs = ref([])
    
    // 服务控制输出
    const serviceOutput = ref('')
    
    // 状态标志
    const isInstalling = ref(false)
    
    // 日志容器引用
    const logContainer = ref(null)
    
    // 定时器
    let statusInterval = null
    let logsInterval = null
    
    // 格式化日志显示
    const formatLogs = (logs) => {
      if (!logs || logs.length === 0) return 'No logs available'
      return logs.join('\n')
    }
    
    // 获取Slurm配置
    const loadConfig = async () => {
      try {
        const config = await fetchSlurmConfig()
        slurmConfig.value = {
          ...config,
          compute_nodes: config.compute_nodes ? config.compute_nodes.join(', ') : ''
        }
      } catch (error) {
        console.error('Failed to load Slurm configuration:', error)
      }
    }
    
    // 更新配置
    const updateConfiguration = async () => {
      try {
        const configToSend = {
          ...slurmConfig.value,
          compute_nodes: slurmConfig.value.compute_nodes
            .split(',')
            .map(node => node.trim())
            .filter(node => node.length > 0)
        }
        
        await updateSlurmConfig(configToSend)
        alert('Configuration updated successfully')
      } catch (error) {
        alert('Failed to update configuration: ' + error.message)
      }
    }
    
    // 获取部署状态
    const loadDeploymentStatus = async () => {
      try {
        deploymentStatus.value = await fetchDeploymentStatus()
      } catch (error) {
        console.error('Failed to load deployment status:', error)
      }
    }
    
    // 启动安装
    const startInstallation = async () => {
      try {
        isInstalling.value = true
        await startDeployment()
        // 自动跳转到日志选项卡
        activeTab.value = 2
        // 启动状态轮询
        startStatusPolling()
      } catch (error) {
        alert('Failed to start deployment: ' + error.message)
      } finally {
        isInstalling.value = false
      }
    }
    
    // 获取部署日志
    const loadDeploymentLogs = async () => {
      try {
        deploymentLogs.value = await fetchDeploymentLogs()
        // 滚动到底部
        await nextTick()
        if (logContainer.value && logContainer.value.$el) {
          const textarea = logContainer.value.$el.querySelector('textarea')
          if (textarea) {
            textarea.scrollTop = textarea.scrollHeight
          }
        }
      } catch (error) {
        console.error('Failed to load deployment logs:', error)
      }
    }
    
    // 刷新日志
    const refreshLogs = async () => {
      await loadDeploymentLogs()
    }
    
    // 控制服务
    const controlService = async (action) => {
      try {
        const result = await controlSlurmService(action)
        serviceOutput.value = result.output || 'Command executed successfully'
        // 更新状态
        checkSlurmStatus()
      } catch (error) {
        serviceOutput.value = 'Error: ' + error.message
      }
    }
    
    // 获取Slurm状态
    const loadSlurmStatus = async () => {
      try {
        slurmStatus.value = await fetchSlurmStatus()
      } catch (error) {
        console.error('Failed to load Slurm status:', error)
      }
    }
    
    // 启动状态轮询
    const startStatusPolling = () => {
      // 清除现有定时器
      if (statusInterval) clearInterval(statusInterval)
      if (logsInterval) clearInterval(logsInterval)
      
      // 设置新的定时器
      statusInterval = setInterval(async () => {
        await loadDeploymentStatus()
        await loadSlurmStatus()
      }, 3000)
      
      logsInterval = setInterval(loadDeploymentLogs, 5000)
    }
    
    // 停止状态轮询
    const stopStatusPolling = () => {
      if (statusInterval) {
        clearInterval(statusInterval)
        statusInterval = null
      }
      if (logsInterval) {
        clearInterval(logsInterval)
        logsInterval = null
      }
    }
    
    // 组件挂载时加载数据
    onMounted(async () => {
      await loadConfig()
      await loadDeploymentStatus()
      await loadDeploymentLogs()
      await loadSlurmStatus()
    })
    
    // 组件卸载时清理定时器
    onUnmounted(() => {
      stopStatusPolling()
    })
    
    return {
      slurmConfig,
      activeTab,
      slurmStatus,
      deploymentStatus,
      deploymentLogs,
      serviceOutput,
      isInstalling,
      logContainer,
      formatLogs,
      updateConfiguration,
      startInstallation,
      refreshLogs,
      controlService,
      loadSlurmStatus
    }
  }
}
</script>

<style scoped>
.v-card {
  transition: all 0.3s ease;
}

.v-btn {
  text-transform: none;
}

:deep(.v-alert__content) strong {
  color: white;
}

:deep(.v-alert--outlined.success strong) {
  color: #4caf50;
}

:deep(.v-alert--outlined.error strong) {
  color: #f44336;
}
</style>