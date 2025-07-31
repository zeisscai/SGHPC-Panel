<template>
  <div class="card">
    <h2>部署控制</h2>
    <div>
      <button @click="startDeployment" class="btn btn-success" :disabled="deploying">
        {{ deploying ? '部署中...' : '开始部署' }}
      </button>
      <button @click="stopDeployment" class="btn btn-danger" :disabled="!deploying">
        停止部署
      </button>
    </div>
    
    <div style="margin-top: 1rem;">
      <h3>部署选项</h3>
      <div>
        <label>
          <input type="checkbox" v-model="options.cleanupPrevious">
          清理之前的安装
        </label>
      </div>
      <div>
        <label>
          <input type="checkbox" v-model="options.useUSTCRepo">
          使用USTC软件源
        </label>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DeploymentControl',
  data() {
    return {
      deploying: false,
      options: {
        cleanupPrevious: true,
        useUSTCRepo: true
      }
    }
  },
  methods: {
    startDeployment() {
      this.deploying = true;
      this.$emit('start-deployment', this.options);
    },
    
    stopDeployment() {
      this.deploying = false;
      this.$emit('stop-deployment');
    }
  }
}
</script>