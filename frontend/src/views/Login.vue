<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="6" lg="4">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>HPC Control Panel Login</v-toolbar-title>
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
            </v-form>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn 
              color="primary" 
              @click="login" 
              :disabled="!valid || loading"
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
import { login } from '@/api/node'

export default {
  name: 'Login',
  setup() {
    const router = useRouter()
    const valid = ref(false)
    const loading = ref(false)
    const username = ref('')
    const password = ref('')

    const usernameRules = [
      v => !!v || 'Username is required',
      v => v.length >= 3 || 'Username must be at least 3 characters'
    ]

    const passwordRules = [
      v => !!v || 'Password is required',
      v => v.length >= 6 || 'Password must be at least 6 characters'
    ]

    const loginHandler = async () => {
      if (!valid.value) return
      
      loading.value = true
      
      try {
        const data = await login(username.value, password.value)
        
        // 保存认证信息到localStorage
        localStorage.setItem('authToken', data.token)
        localStorage.setItem('user', JSON.stringify(data.user))
        localStorage.setItem('lastActivity', Date.now().toString())
        
        // 检查是否使用默认密码
        if (data.is_default_password) {
          // 设置标志，提醒用户修改密码
          localStorage.setItem('shouldChangePassword', 'true')
        }
        
        // 重定向到Overview页面
        router.push('/')
      } catch (error) {
        console.error('Login error:', error)
        alert('Invalid username or password')
      } finally {
        loading.value = false
      }
    }

    return {
      valid,
      loading,
      username,
      password,
      usernameRules,
      passwordRules,
      login: loginHandler
    }
  }
}
</script>

<style scoped>
.fill-height {
  min-height: 100vh;
}
</style>