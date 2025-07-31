<template>
  <div id="app">
    <header>
      <h1>Slurm 集群部署面板</h1>
    </header>
    
    <main>
      <div class="container">
        <DeploymentForm @add-node="addNode" @load-sample="loadSampleConfig" @clear-config="clearConfig" />
        <NodeList :nodes="nodes" @remove-node="removeNode" />
        <DeploymentControl @start-deployment="startDeployment" @stop-deployment="stopDeployment" />
        <LogViewer :logs="logs" @clear-logs="clearLogs" @download-logs="downloadLogs" />
      </div>
    </main>
    
    <footer>
      <p>Slurm 部署面板 v1.0</p>
    </footer>
  </div>
</template>

<script>
import DeploymentForm from './DeploymentForm.vue'
import NodeList from './NodeList.vue'
import DeploymentControl from './DeploymentControl.vue'
import LogViewer from './LogViewer.vue'

export default {
  name: 'App',
  components: {
    DeploymentForm,
    NodeList,
    DeploymentControl,
    LogViewer
  },
  data() {
    return {
      nodes: [],
      logs: ''
    }
  },
  methods: {
    addNode(node) {
      // 为节点添加ID和默认状态
      const newNode = {
        ...node,
        id: Date.now(),
        status: '未部署'
      };
      this.nodes.push(newNode);
    },
    
    removeNode(index) {
      this.nodes.splice(index, 1);
    },
    
    loadSampleConfig() {
      // 加载示例配置
      this.nodes = [
        {
          id: 1,
          type: 'master',
          ip: '192.168.1.100',
          hostname: 'master',
          password: 'password',
          status: '未部署'
        },
        {
          id: 2,
          type: 'compute',
          ip: '192.168.1.101',
          hostname: 'compute1',
          password: 'password',
          status: '未部署'
        }
      ];
    },
    
    clearConfig() {
      this.nodes = [];
    },
    
    startDeployment(options) {
      console.log('开始部署:', options);
      // 这里应该调用后端API开始部署
    },
    
    stopDeployment() {
      console.log('停止部署');
      // 这里应该调用后端API停止部署
    },
    
    clearLogs() {
      this.logs = '';
    },
    
    downloadLogs() {
      console.log('下载日志');
      // 实现日志下载功能
    }
  }
}
</script>

<style>
body {
  font-family: Arial, sans-serif;
  margin: 0;
  padding: 0;
  background-color: #f5f5f5;
}

header {
  background-color: #2c3e50;
  color: white;
  padding: 1rem;
}

header h1 {
  margin: 0;
}

.container {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-bottom: 2rem;
  padding: 1.5rem;
}

.card h2 {
  margin-top: 0;
  color: #2c3e50;
}

.btn {
  background-color: #3498db;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn:hover {
  background-color: #2980b9;
}

.btn-danger {
  background-color: #e74c3c;
}

.btn-danger:hover {
  background-color: #c0392b;
}

.btn-success {
  background-color: #27ae60;
}

.btn-success:hover {
  background-color: #229954;
}

form div {
  margin-bottom: 1rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
}

input, select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

th {
  background-color: #f8f9fa;
  font-weight: bold;
}
</style>