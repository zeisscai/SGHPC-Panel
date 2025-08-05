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
            <v-alert
              v-if="deploymentStatus.phase === 'failed'"
              type="error"
              outlined
            >
              <strong>Deployment Failed:</strong> {{ deploymentStatus.error_message }}
            </v-alert>
            
            <v-alert
              v-if="deploymentStatus.phase === 'finished'"
              type="success"
              outlined
            >
              <strong>Deployment Completed Successfully!</strong>
            </v-alert>
            
            <v-row>
              <v-col cols="12" md="6">
                <v-card outlined>
                  <v-card-title>Deployment Status</v-card-title>
                  <v-card-text>
                    <v-progress-linear
                      :value="deploymentStatus.progress"
                      :color="getProgressColor(deploymentStatus.phase)"
                      height="25"
                      striped
                    >
                      <strong>{{ deploymentStatus.progress }}%</strong>
                    </v-progress-linear>
                    
                    <v-list>
                      <v-list-item>
                        <v-list-item-title>Phase</v-list-item-title>
                        <v-list-item-subtitle>{{ deploymentStatus.phase }}</v-list-item-subtitle>
                      </v-list-item>
                      <v-list-item>
                        <v-list-item-title>Message</v-list-item-title>
                        <v-list-item-subtitle>{{ deploymentStatus.message }}</v-list-item-subtitle>
                      </v-list-item>
                      <v-list-item>
                        <v-list-item-title>Start Time</v-list-item-title>
                        <v-list-item-subtitle>{{ deploymentStatus.start_time }}</v-list-item-subtitle>
                      </v-list-item>
                      <v-list-item v-if="deploymentStatus.end_time">
                        <v-list-item-title>End Time</v-list-item-title>
                        <v-list-item-subtitle>{{ deploymentStatus.end_time }}</v-list-item-subtitle>
                      </v-list-item>
                    </v-list>
                    
                    <v-btn
                      color="primary"
                      @click="startDeployment"
                      :disabled="isDeploying"
                      :loading="isDeploying"
                    >
                      <v-icon left>mdi-rocket</v-icon>
                      Start Deployment
                    </v-btn>
                  </v-card-text>
                </v-card>
              </v-col>
              
              <v-col cols="12" md="6">
                <v-card outlined>
                  <v-card-title>Service Control</v-card-title>
                  <v-card-text>
                    <v-row>
                      <v-col cols="12">
                        <v-btn
                          color="success"
                          @click="controlService('start')"
                          class="mr-2 mb-2"
                        >
                          <v-icon left>mdi-play</v-icon>
                          Start
                        </v-btn>
                        <v-btn
                          color="warning"
                          @click="controlService('restart')"
                          class="mr-2 mb-2"
                        >
                          <v-icon left>mdi-restart</v-icon>
                          Restart
                        </v-btn>
                        <v-btn
                          color="error"
                          @click="controlService('stop')"
                          class="mr-2 mb-2"
                        >
                          <v-icon left>mdi-stop</v-icon>
                          Stop
                        </v-btn>
                      </v-col>
                    </v-row>
                    
                    <v-row>
                      <v-col cols="12">
                        <v-textarea
                          outlined
                          label="Service Control Output"
                          :model-value="serviceOutput"
                          rows="5"
                          readonly
                        ></v-textarea>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
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
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon left class="mr-2">mdi-file-document</v-icon>
            Deployment Logs
          </v-card-title>
          <v-card-text>
            <v-textarea
              outlined
              :model-value="formatLogs(deploymentLogs)"
              rows="10"
              readonly
              no-resize
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
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  fetchSlurmConfig, 
  updateSlurmConfig, 
  fetchDeploymentStatus, 
  startDeployment,
  fetchDeploymentLogs,
  controlSlurmService
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
    const isDeploying = ref(false)
    
    // 定时器
    let statusInterval = null
    let logsInterval = null
    
    // 获取进度条颜色
    const getProgressColor = (phase) => {
      switch (phase) {
        case 'failed':
          return 'error'
        case 'finished':
          return 'success'
        case 'checking':
        case 'downloading':
        case 'installing_deps':
        case 'compiling':
        case 'configuring':
          return 'info'
        default:
          return 'primary'
      }
    }
    
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
    
    // 启动部署
    const startDeploymentProcess = async () => {
      try {
        isDeploying.value = true
        await startDeployment()
        alert('Deployment started successfully')
        // 启动状态轮询
        startStatusPolling()
      } catch (error) {
        alert('Failed to start deployment: ' + error.message)
      } finally {
        isDeploying.value = false
      }
    }
    
    // 获取部署日志
    const loadDeploymentLogs = async () => {
      try {
        deploymentLogs.value = await fetchDeploymentLogs()
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
      } catch (error) {
        serviceOutput.value = 'Error: ' + error.message
      }
    }
    
    // 启动状态轮询
    const startStatusPolling = () => {
      // 清除现有定时器
      if (statusInterval) clearInterval(statusInterval)
      if (logsInterval) clearInterval(logsInterval)
      
      // 设置新的定时器
      statusInterval = setInterval(loadDeploymentStatus, 3000)
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
    })
    
    // 组件卸载时清理定时器
    onUnmounted(() => {
      stopStatusPolling()
    })
    
    return {
      slurmConfig,
      deploymentStatus,
      deploymentLogs,
      serviceOutput,
      isDeploying,
      getProgressColor,
      formatLogs,
      updateConfiguration,
      startDeployment: startDeploymentProcess,
      refreshLogs,
      controlService
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
</style>