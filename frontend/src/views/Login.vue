<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="6" lg="4">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>SGHPC Panel Login</v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-alert
              type="info"
              outlined
              class="mb-4"
            >
              <div class="font-weight-bold">Default Credentials:</div>
              <div>Username: <strong>admin</strong></div>
              <div>Password: <strong>password</strong></div>
            </v-alert>
            
            <v-form ref="loginForm" v-model="valid" @submit.prevent="login">
              <v-text-field
                v-model="username"
                label="Username"
                name="username"
                prepend-icon="mdi-account"
                type="text"
                :rules="usernameRules"
                required
              ></v-text-field>

              <v-text-field
                v-model="password"
                label="Password"
                name="password"
                prepend-icon="mdi-lock"
                type="password"
                :rules="passwordRules"
                required
              ></v-text-field>
              
              <v-checkbox
                v-model="rememberMe"
                label="Remember me"
                class="mt-2"
              ></v-checkbox>
            </v-form>
          </v-card-text>
          <v-card-actions class="pa-6">
            <v-spacer></v-spacer>
            <v-btn 
              color="primary" 
              @click="login"
              :disabled="!valid"
              :loading="loading"
            >
              Login
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

export default {
  name: 'Login',
  setup() {
    const router = useRouter()
    const valid = ref(false)
    const loading = ref(false)
    const username = ref('')
    const password = ref('')
    const rememberMe = ref(false)
    
    const usernameRules = [
      v => !!v || 'Username is required'
    ]
    
    const passwordRules = [
      v => !!v || 'Password is required'
    ]
    
    const login = async () => {
      if (!valid.value) return
      
      loading.value = true
      
      try {
        // 发送登录请求到后端
        const response = await axios.post('/api/login', {
          username: username.value,
          password: password.value
        })
        
        // 保存认证信息
        const { token, user } = response.data
        localStorage.setItem('authToken', token)
        localStorage.setItem('user', JSON.stringify(user))
        localStorage.setItem('lastActivity', Date.now().toString())
        
        // 检查是否需要更改密码
        if (user.shouldChangePassword) {
          localStorage.setItem('shouldChangePassword', 'true')
        }
        
        // 如果选择了"记住我"，设置更长的超时时间
        const timeout = rememberMe.value ? 7 * 24 * 60 * 60 * 1000 : 5 * 60 * 1000
        localStorage.setItem('sessionTimeout', timeout.toString())
        
        // 跳转到主页面
        router.push('/')
      } catch (error) {
        // 显示错误消息
        alert('Login failed: ' + (error.response?.data?.message || 'Invalid credentials'))
      } finally {
        loading.value = false
      }
    }
    
    return {
      valid,
      loading,
      username,
      password,
      rememberMe,
      usernameRules,
      passwordRules,
      login
    }
  }
}
</script>

<style scoped>
.v-card {
  border-radius: 12px !important;
}

.v-toolbar {
  border-top-left-radius: 12px !important;
  border-top-right-radius: 12px !important;
}
</style>