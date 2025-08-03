<template>
  <div class="settings">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon start class="mr-2">mdi-cog</v-icon>
            Settings
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12" md="6">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon start class="mr-2">mdi-account-key</v-icon>
            Change Password
          </v-card-title>
          <v-card-text>
            <v-form ref="passwordForm" v-model="valid">
              <v-text-field
                v-model="currentPassword"
                label="Current Password"
                type="password"
                :rules="currentPasswordRules"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-text-field
                v-model="newPassword"
                label="New Password"
                type="password"
                :rules="newPasswordRules"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-text-field
                v-model="confirmPassword"
                label="Confirm New Password"
                type="password"
                :rules="confirmPasswordRules"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-btn 
                color="primary" 
                @click="changePassword" 
                :disabled="!valid || loading"
                :loading="loading"
              >
                Change Password
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
      
      <v-col cols="12" md="6">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon start class="mr-2">mdi-information</v-icon>
            User Information
          </v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item>
                <v-list-item-title>Username</v-list-item-title>
                <v-list-item-subtitle>{{ user.username }}</v-list-item-subtitle>
              </v-list-item>
              
              <v-list-item>
                <v-list-item-title>Last Login</v-list-item-title>
                <v-list-item-subtitle>{{ lastLogin }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { changePassword } from '@/api/node'

export default {
  name: 'Settings',
  setup() {
    const valid = ref(false)
    const loading = ref(false)
    const currentPassword = ref('')
    const newPassword = ref('')
    const confirmPassword = ref('')
    
    // 从localStorage获取用户信息
    const user = computed(() => {
      const userData = localStorage.getItem('user')
      return userData ? JSON.parse(userData) : { username: 'Unknown' }
    })
    
    const lastLogin = ref(new Date().toLocaleString())
    
    const currentPasswordRules = [
      v => !!v || 'Current password is required'
    ]
    
    const newPasswordRules = [
      v => !!v || 'New password is required',
      v => v.length >= 6 || 'Password must be at least 6 characters',
      v => v !== currentPassword.value || 'New password must be different from current password'
    ]
    
    const confirmPasswordRules = [
      v => !!v || 'Please confirm your new password',
      v => v === newPassword.value || 'Passwords do not match'
    ]
    
    const changePasswordHandler = async () => {
      if (!valid.value) return
      
      loading.value = true
      
      try {
        await changePassword(currentPassword.value, newPassword.value)
        
        // 重置表单
        currentPassword.value = ''
        newPassword.value = ''
        confirmPassword.value = ''
        
        alert('Password changed successfully!')
      } catch (error) {
        console.error('Password change error:', error)
        alert('Failed to change password. Please check your current password and try again.')
      } finally {
        loading.value = false
      }
    }
    
    return {
      valid,
      loading,
      currentPassword,
      newPassword,
      confirmPassword,
      user,
      lastLogin,
      currentPasswordRules,
      newPasswordRules,
      confirmPasswordRules,
      changePassword: changePasswordHandler
    }
  }
}
</script>

<style scoped>
.settings {
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.v-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.v-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.2) !important;
}
</style>