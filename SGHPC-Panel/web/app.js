const { createApp } = Vue

// Configuration Form Component
const ConfigurationForm = {
  data() {
    return {
      nodes: [
        { name: 'master', ip: '', password: '', hostname: '' },
        { name: 'node1', ip: '', password: '', hostname: '' },
        { name: 'node2', ip: '', password: '', hostname: '' },
        { name: 'node3', ip: '', password: '', hostname: '' },
        { name: 'node4', ip: '', password: '', hostname: '' }
      ],
      loading: false
    }
  },
  mounted() {
    this.loadConfig();
  },
  methods: {
    async loadConfig() {
      try {
        const response = await fetch('/api/config');
        const data = await response.json();
        if (data.nodes) {
          this.nodes = data.nodes;
        }
      } catch (error) {
        console.error('Error loading config:', error);
      }
    },
    async saveConfig() {
      this.loading = true;
      try {
        const response = await fetch('/api/config', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ nodes: this.nodes })
        });
        
        if (response.ok) {
          alert('Configuration saved successfully!');
        } else {
          alert('Error saving configuration');
        }
      } catch (error) {
        console.error('Error saving config:', error);
        alert('Error saving configuration');
      } finally {
        this.loading = false;
      }
    }
  },
  template: `
    <div class="mt-4">
      <h3>Node Configuration</h3>
      <p>Configure your cluster nodes. At least master node is required.</p>
      
      <div class="row">
        <div class="col-md-12" v-for="(node, index) in nodes" :key="node.name">
          <div class="card node-card">
            <div class="card-header">
              <strong>{{ node.name.toUpperCase() }}</strong>
            </div>
            <div class="card-body">
              <div class="row">
                <div class="col-md-3">
                  <label :for="'ip-' + index" class="form-label">IP Address</label>
                  <input type="text" class="form-control" :id="'ip-' + index" v-model="node.ip" placeholder="e.g. 192.168.1.10">
                </div>
                <div class="col-md-3">
                  <label :for="'password-' + index" class="form-label">Password</label>
                  <input type="password" class="form-control" :id="'password-' + index" v-model="node.password">
                </div>
                <div class="col-md-3">
                  <label :for="'hostname-' + index" class="form-label">Hostname</label>
                  <input type="text" class="form-control" :id="'hostname-' + index" v-model="node.hostname" :placeholder="node.name">
                </div>
                <div class="col-md-3">
                  <label class="form-label">Status</label>
                  <div>
                    <span v-if="node.ip" class="badge bg-success">Configured</span>
                    <span v-else class="badge bg-secondary">Not configured</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="row">
        <div class="col-12">
          <button class="btn btn-primary" @click="saveConfig" :disabled="loading">
            <span v-if="loading" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
            Save Configuration
          </button>
        </div>
      </div>
    </div>
  `
}

// Deployment Panel Component
const DeploymentPanel = {
  data() {
    return {
      deploying: false,
      status: {
        running: false,
        message: 'Ready to deploy',
        completed: false
      },
      interval: null
    }
  },
  mounted() {
    this.checkStatus();
  },
  beforeUnmount() {
    if (this.interval) {
      clearInterval(this.interval);
    }
  },
  methods: {
    async startDeployment() {
      if (!confirm('Are you sure you want to start the deployment?')) {
        return;
      }
      
      this.deploying = true;
      try {
        const response = await fetch('/api/deploy', {
          method: 'POST'
        });
        
        if (response.ok) {
          this.status.message = 'Deployment started...';
          this.startStatusPolling();
        } else {
          alert('Error starting deployment');
        }
      } catch (error) {
        console.error('Error starting deployment:', error);
        alert('Error starting deployment');
      } finally {
        this.deploying = false;
      }
    },
    startStatusPolling() {
      this.interval = setInterval(() => {
        this.checkStatus();
      }, 2000);
    },
    async checkStatus() {
      try {
        const response = await fetch('/api/status');
        const data = await response.json();
        this.status = data;
        
        if (this.status.completed && this.interval) {
          clearInterval(this.interval);
          this.interval = null;
        }
      } catch (error) {
        console.error('Error checking status:', error);
      }
    }
  },
  template: `
    <div class="mt-4">
      <h3>Deployment</h3>
      <p>Start the Slurm deployment process on your configured nodes.</p>
      
      <div class="row">
        <div class="col-md-12">
          <div class="alert alert-info">
            <h5>Before Deployment</h5>
            <ul>
              <li>Ensure all nodes are configured in the Configuration tab</li>
              <li>Make sure all nodes are accessible via SSH</li>
              <li>Verify that the deploy.conf file is correctly generated</li>
            </ul>
          </div>
        </div>
      </div>
      
      <div class="row">
        <div class="col-md-12">
          <button class="btn btn-success btn-lg" @click="startDeployment" :disabled="deploying || status.running">
            <span v-if="deploying" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
            Start Deployment
          </button>
        </div>
      </div>
      
      <div class="row">
        <div class="col-md-12">
          <div class="status-box">
            <h5>Deployment Status</h5>
            <p>{{ status.message }}</p>
            <div v-if="status.running" class="progress">
              <div class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style="width: 100%"></div>
            </div>
            <div v-if="status.completed && !status.running" class="alert alert-success mt-2">
              Deployment process completed!
            </div>
          </div>
        </div>
      </div>
      
      <div class="row mt-4">
        <div class="col-md-12">
          <div class="card">
            <div class="card-header">
              <strong>Deployment Information</strong>
            </div>
            <div class="card-body">
              <p>This deployment panel automates the installation of Slurm 24.11.5 on Rocky Linux 9.6 using the OpenHPC repository.</p>
              <p>The process includes:</p>
              <ul>
                <li>Setting up SSH connections to all nodes</li>
                <li>Configuring hosts files across the cluster</li>
                <li>Installing required packages via OpenHPC</li>
                <li>Setting up Munge authentication</li>
                <li>Configuring MariaDB for accounting</li>
                <li>Deploying and starting Slurm services</li>
                <li>Verifying the installation</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  `
}

// Main App
const app = createApp({})

// Register components
app.component('configuration-form', ConfigurationForm)
app.component('deployment-panel', DeploymentPanel)

// Mount the app
app.mount('#app')