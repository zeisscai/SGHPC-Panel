<template>
  <div class="overview">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            Status Overview
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2" transition="slide-y-transition">
          <v-card-title>
            <v-icon left class="mr-2">mdi-server</v-icon>
            Management Node Information
          </v-card-title>
          <v-card-text>
            <v-table v-if="managementNode" density="comfortable">
              <tbody>
                <tr>
                  <td class="font-weight-bold">Hostname</td>
                  <td>{{ managementNode.hostname }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Model</td>
                  <td>{{ managementNode.model }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Architecture</td>
                  <td>{{ managementNode.architecture }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">CPU Information</td>
                  <td>{{ managementNode.cpu_info }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">OS Version</td>
                  <td>{{ managementNode.os_version }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Kernel Version</td>
                  <td>{{ managementNode.kernel_version }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Local Time</td>
                  <td>{{ managementNode.local_time }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Uptime</td>
                  <td>{{ managementNode.uptime }}</td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2" transition="slide-y-transition">
          <v-card-title>
            <v-icon left class="mr-2">mdi-server-network</v-icon>
            Compute Nodes
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="nodeHeaders"
              :items="computeNodes"
              class="elevation-1"
              :loading="loadingNodes"
              hide-default-footer
            >
              <template v-slot:item.cpu_usage="{ item }">
                <v-chip 
                  :color="getUsageColor(item.cpu_usage)" 
                  dark
                >
                  {{ Math.round(item.cpu_usage) }}%
                </v-chip>
              </template>
              <template v-slot:item.memory_usage="{ item }">
                <v-chip 
                  :color="getUsageColor(item.memory_usage)" 
                  dark
                >
                  {{ (item.memory_usage * 0.16).toFixed(1) }}GB/16GB
                </v-chip>
              </template>
              <template v-slot:item.status="{ item }">
                <v-chip 
                  :color="item.status === 'High Load' ? 'error' : 'success'" 
                  dark
                >
                  {{ item.status }}
                </v-chip>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2" transition="slide-y-transition">
          <v-card-title>
            <v-icon left class="mr-2">mdi-format-list-bulleted</v-icon>
            SLURM Jobs
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="jobHeaders"
              :items="activeJobs"
              class="elevation-1"
              :loading="loadingJobs"
              hide-default-footer
            >
              <template v-slot:item.submission_time="{ item }">
                {{ new Date(item.submission_time).toLocaleString() }}
              </template>
              <template v-slot:item.status="{ item }">
                <v-chip 
                  :color="getStatusColor(item.status)" 
                  dark
                >
                  {{ item.status }}
                </v-chip>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { fetchManagementNode, fetchComputeNodes, fetchSlurmJobs } from '../api/node'

export default {
  name: 'Overview',
  setup() {
    const managementNode = ref(null)
    const computeNodes = ref([])
    const slurmJobs = ref([])
    const loadingNodes = ref(false)
    const loadingJobs = ref(false)
    
    // 只显示等待或正在计算的作业（过滤掉已完成、取消的作业）
    const activeJobs = computed(() => {
      return slurmJobs.value.filter(job => job.status === 'pending' || job.status === 'running')
    })
    
    const nodeHeaders = [
      { title: 'Hostname', key: 'hostname' },
      { title: 'IP Address', key: 'ip' },
      { title: 'CPU Usage', key: 'cpu_usage' }, // 修改标题，去掉百分比符号
      { title: 'Memory Usage', key: 'memory_usage' }, // 修改标题
      { title: 'Status', key: 'status' }
    ]
    
    const jobHeaders = [
      { title: 'Job ID', key: 'job_id' },
      { title: 'Submit Time', key: 'submission_time' },
      { title: 'Wait Time', key: 'wait_time' },
      { title: 'Compute Time', key: 'compute_time' },
      { title: 'User', key: 'user' },
      { title: 'Status', key: 'status' }
    ]
    
    const getUsageColor = (value) => {
      if (value > 80) return 'error'
      if (value > 60) return 'warning'
      return 'success'
    }
    
    const getStatusColor = (status) => {
      switch (status) {
        case 'pending':
          return 'warning'
        case 'running':
          return 'info'
        case 'completed':
          return 'success'
        case 'cancelled':
          return 'error'
        default:
          return 'default'
      }
    }
    
    const loadManagementNode = async () => {
      try {
        managementNode.value = await fetchManagementNode()
      } catch (error) {
        console.error('Failed to load management node:', error)
      }
    }
    
    const loadComputeNodes = async () => {
      loadingNodes.value = true
      try {
        const nodes = await fetchComputeNodes()
        // 添加状态字段
        computeNodes.value = nodes.map(node => ({
          ...node,
          status: node.cpu_usage > 80 ? 'High Load' : 'Normal'
        }))
      } catch (error) {
        console.error('Failed to load compute nodes:', error)
      } finally {
        loadingNodes.value = false
      }
    }
    
    const loadSlurmJobs = async () => {
      loadingJobs.value = true
      try {
        slurmJobs.value = await fetchSlurmJobs()
      } catch (error) {
        console.error('Failed to load SLURM jobs:', error)
      } finally {
        loadingJobs.value = false
      }
    }
    
    onMounted(() => {
      loadManagementNode()
      loadComputeNodes()
      loadSlurmJobs()
    })
    
    return {
      managementNode,
      computeNodes,
      slurmJobs,
      activeJobs,
      nodeHeaders,
      jobHeaders,
      loadingNodes,
      loadingJobs,
      getUsageColor,
      getStatusColor,
      Math // 添加Math对象以便在模板中使用
    }
  }
}
</script>

<style scoped>
.v-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.v-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.2) !important;
}

.v-data-table {
  border-radius: 8px;
}

.v-table tbody tr:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.v-table tbody td {
  padding: 12px 16px;
}

.v-table tbody td:first-child {
  width: 30%;
  font-weight: 500;
}
</style>