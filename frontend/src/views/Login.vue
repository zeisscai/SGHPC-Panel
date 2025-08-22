<template>
  <v-container class="fill-height login-container" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="10" md="8" lg="5">
        <v-card class="login-card elevation-12">
          <v-toolbar color="primary" dark flat class="toolbar">
            <v-toolbar-title class="text-h5 font-weight-bold text-center w-100">
              SGHPC Panel Login
            </v-toolbar-title>
          </v-toolbar>
          
          <v-card-text class="pa-8">
            <v-form ref="loginForm" v-model="valid" @submit.prevent="login" autocomplete="off">
              <v-text-field
                v-model="username"
                label="Username"
                name="username"
                prepend-inner-icon="mdi-account"
                type="text"
                :rules="usernameRules"
                required
                @keyup.enter="login"
                variant="outlined"
                class="mb-4"
                clearable
              ></v-text-field>

              <v-text-field
                v-model="password"
                label="Password"
                name="password"
                prepend-inner-icon="mdi-lock"
                :type="showPassword ? 'text' : 'password'"
                :rules="passwordRules"
                required
                @keyup.enter="login"
                :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append-inner="showPassword = !showPassword"
                variant="outlined"
                class="mb-4"
                clearable
              ></v-text-field>
              
              <v-checkbox
                v-model="rememberMe"
                label="Remember me"
                color="primary"
                class="mb-6"
                hide-details
              ></v-checkbox>
              
              <v-btn 
                color="primary" 
                @click="login"
                :disabled="!valid || loading"
                :loading="loading"
                block
                size="large"
                class="login-btn"
                rounded="lg"
              >
                <span class="text-body-1 font-weight-bold">Login</span>
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
        
        <div class="text-center text-caption text-medium-emphasis mt-6">
          © {{ new Date().getFullYear() }} SGHPC Panel - SG-HPC Inc.
        </div>
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
    const loginForm = ref(null)
    const formErrors = ref([])
    const router = useRouter()
    const valid = ref(false)
    const loading = ref(false)
    const username = ref('')
    const password = ref('')
    const rememberMe = ref(false)
    const showPassword = ref(false)
    
    const usernameRules = [
      v => !!v || 'Username is required',
      v => (v && v.length <= 50) || 'Username must be less than 50 characters'
    ]
    
    const passwordRules = [
      v => !!v || 'Password is required'
    ]

    const passwordErrors = ref([])
    
    const clearPasswordError = () => {
      if (passwordErrors.value.length > 0) {
        passwordErrors.value = []
      }
    }
    
    const login = async () => {
      // 强制触发表单验证
      const { valid } = await loginForm.value.validate()
      if (!valid) return
      
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
        if (response.data.is_default_password) {
          localStorage.setItem('shouldChangePassword', 'true')
          // 跳转到修改密码页面
          router.push('/change-password')
        } else {
          // 跳转到主页面
          router.push('/')
        }
        
        // 如果选择了"记住我"，设置更长的超时时间
        const timeout = rememberMe.value ? 7 * 24 * 60 * 60 * 1000 : 5 * 60 * 1000
        localStorage.setItem('sessionTimeout', timeout.toString())
      } catch (error) {
        // 显示错误消息
        let errorMsg = '';
        if (error.response) {
          // 服务器返回错误响应
          if (error.response.status === 401) {
            errorMsg = 'Invalid username or password';
            passwordErrors.value = [errorMsg]
          } else if (error.response.status === 400) {
            errorMsg = 'Invalid request';
          } else if (error.response.status === 500) {
            errorMsg = 'Server error. Please try again later.';
          } else {
            errorMsg = error.response.data || 'Unknown error occurred';
          }
        } else if (error.request) {
          // 请求发出但没有收到响应
          errorMsg = 'No response from server. Please check your connection and try again.';
        } else {
          // 其他错误
          errorMsg = error.message || 'Unknown error occurred';
        }
        
        // 创建错误提示框
        const errorDiv = document.createElement('div')
        errorDiv.style.position = 'fixed'
        errorDiv.style.top = '20px'
        errorDiv.style.right = '20px'
        errorDiv.style.backgroundColor = '#ff5252'
        errorDiv.style.color = 'white'
        errorDiv.style.padding = '15px 25px'
        errorDiv.style.borderRadius = '4px'
        errorDiv.style.boxShadow = '0 2px 10px rgba(0,0,0,0.2)'
        errorDiv.style.zIndex = '9999'
        errorDiv.style.transition = 'opacity 0.3s'
        errorDiv.style.opacity = '1'
        errorDiv.innerText = errorMsg
        
        document.body.appendChild(errorDiv)
        
        // 3秒后淡出并移除提示框
        setTimeout(() => {
          errorDiv.style.opacity = '0'
          setTimeout(() => {
            if (errorDiv.parentNode) {
              document.body.removeChild(errorDiv)
            }
          }, 300)
        }, 3000)
      } finally {
        loading.value = false
      }
    }
    
    return {
      loginForm,
      valid,
      loading,
      username,
      password,
      rememberMe,
      showPassword,
      usernameRules,
      passwordRules,
      passwordErrors,
      clearPasswordError,
      login
    }
  }
}
</script>

<style scoped>
.login-container {
  background-color: #121212;
  min-height: 100vh;
  display: flex;
  align-items: center;
}

.login-card {
  border-radius: 16px !important;
  overflow: hidden;
  background: #1e1e1e;
}

.toolbar {
  border-top-left-radius: 16px !important;
  border-top-right-radius: 16px !important;
}

.login-btn {
  text-transform: none !important;
  letter-spacing: 1px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.28) !important;
  transition: all 0.3s ease;
}

.login-btn:hover {
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.35) !important;
  transform: translateY(-1px);
}

:deep(.v-field--variant-outlined .v-field__outline__start) {
  border-radius: 8px 0 0 8px !important;
}

:deep(.v-field--variant-outlined .v-field__outline__end) {
  border-radius: 0 8px 8px 0 !important;
}

:deep(.v-input__prepend-inner) {
  margin-right: 12px !important;
}

:deep(.v-input__append-inner) {
  margin-left: 6px !important;
}

@media (max-width: 600px) {
  .login-card {
    margin: 16px;
  }
  
  .pa-8 {
    padding: 24px !important;
  }
}
</style>