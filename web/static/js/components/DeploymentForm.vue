<template>
  <div class="card">
    <h2>集群配置</h2>
    <form @submit.prevent="addNode">
      <div>
        <label for="nodeType">节点类型:</label>
        <select id="nodeType" v-model="newNode.type">
          <option value="master">Master</option>
          <option value="compute">Compute</option>
        </select>
      </div>
      
      <div>
        <label for="nodeIp">IP地址:</label>
        <input type="text" id="nodeIp" v-model="newNode.ip" required>
      </div>
      
      <div>
        <label for="nodeHostname">主机名:</label>
        <input type="text" id="nodeHostname" v-model="newNode.hostname" required>
      </div>
      
      <div>
        <label for="nodePassword">密码:</label>
        <input type="password" id="nodePassword" v-model="newNode.password" required>
      </div>
      
      <button type="submit" class="btn">添加节点</button>
    </form>
    
    <div style="margin-top: 1rem;">
      <button @click="loadSampleConfig" class="btn">加载示例配置</button>
      <button @click="clearConfig" class="btn btn-danger">清空配置</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DeploymentForm',
  data() {
    return {
      newNode: {
        type: 'master',
        ip: '',
        hostname: '',
        password: ''
      }
    }
  },
  methods: {
    addNode() {
      // 添加节点逻辑
      this.$emit('add-node', {...this.newNode});
      
      // 重置表单
      this.newNode = {
        type: 'compute',
        ip: '',
        hostname: '',
        password: ''
      };
    },
    
    loadSampleConfig() {
      // 加载示例配置
      this.$emit('load-sample');
    },
    
    clearConfig() {
      // 清空配置
      this.$emit('clear-config');
    }
  }
}
</script>