<template>
  <div class="settings">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon start class="mr-2 settings-icon">mdi-cog</v-icon>
            Settings
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12" md="6">
        <v-card class="mb-6" elevation="2">
          <v-card-title>
            <v-icon start class="mr-2 settings-icon">mdi-account-key</v-icon>
            Change Password
          </v-card-title>
          <v-card-text>
            <v-alert
              v-if="showPasswordSuccess"
              type="success"
              dismissible
              class="mb-4"
            >
              Password changed successfully!
            </v-alert>
            
            <v-form ref="passwordForm" v-model="passwordFormValid" @submit.prevent="changePassword">
              <v-text-field
                v-model="passwordForm.currentPassword"
                label="Current Password"
                type="password"
                :rules="passwordRules.current"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-text-field
                v-model="passwordForm.newPassword"
                label="New Password"
                type="password"
                :rules="passwordRules.new"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-text-field
                v-model="passwordForm.confirmPassword"
                label="Confirm New Password"
                type="password"
                :rules="passwordRules.confirm"
                required
                class="mb-4"
              ></v-text-field>
              
              <v-btn
                color="primary"
                @click="changePassword"
                :disabled="!passwordFormValid"
                :loading="passwordFormLoading"
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
            <v-icon start class="mr-2 settings-icon">mdi-information</v-icon>
            System Information
          </v-card-title>
          <v-card-text>
            <v-table density="comfortable">
              <tbody>
                <tr>
                  <td class="font-weight-bold">Panel Version</td>
                  <td>{{ systemInfo.version }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Build Date</td>
                  <td>{{ systemInfo.buildDate }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Go Version</td>
                  <td>{{ systemInfo.goVersion }}</td>
                </tr>
                <tr>
                  <td class="font-weight-bold">Operating System</td>
                  <td>{{ systemInfo.os }}</td>
                </tr>
              </tbody>
            </v-table>
            
            <div class="mt-4">
              <v-btn color="warning" @click="restartSystem">
                Restart System
              </v-btn>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import axios from 'axios'

export default {
  name: 'Settings',
  setup() {
    // Password form
    const passwordFormValid = ref(false)
    const passwordFormLoading = ref(false)
    const passwordForm = ref({
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
    
    const passwordRules = ref({
      current: [
        v => !!v || 'Current password is required'
      ],
      new: [
        v => !!v || 'New password is required',
        v => v.length >= 8 || 'Password must be at least 8 characters'
      ],
      confirm: [
        v => !!v || 'Please confirm your new password',
        v => v === passwordForm.value.newPassword || 'Passwords do not match'
      ]
    })
    
    const showPasswordSuccess = ref(false)
    
    // System information
    const systemInfo = ref({
      version: '1.0.0',
      buildDate: '2023-01-01',
      goVersion: 'go1.19.5',
      os: 'Linux'
    })
    
    // Change password
    const changePassword = async () => {
      if (!passwordFormValid.value) return
      
      passwordFormLoading.value = true
      try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        // Reset form
        passwordForm.value = {
          currentPassword: '',
          newPassword: '',
          confirmPassword: ''
        }
        
        // Show success message
        showPasswordSuccess.value = true
        
        // Hide success message after 3 seconds
        setTimeout(() => {
          showPasswordSuccess.value = false
        }, 3000)
        
        // Check if this was the first time password change
        if (localStorage.getItem('shouldChangePassword') === 'true') {
          localStorage.removeItem('shouldChangePassword')
        }
      } catch (error) {
        alert('Failed to change password: ' + (error.response?.data?.message || 'Unknown error'))
      } finally {
        passwordFormLoading.value = false
      }
    }
    
    // Restart system
    const restartSystem = () => {
      if (confirm('Are you sure you want to restart the system?')) {
        // Simulate restart
        alert('System restart initiated')
      }
    }
    
    // Load system information
    onMounted(async () => {
      try {
        // In a real implementation, this would fetch from an API
        // const response = await axios.get('/api/system/info')
        // systemInfo.value = response.data
      } catch (error) {
        console.error('Failed to load system information:', error)
      }
    })
    
    return {
      passwordFormValid,
      passwordFormLoading,
      passwordForm,
      passwordRules,
      showPasswordSuccess,
      systemInfo,
      changePassword,
      restartSystem
    }
  }
}
</script>

<style scoped>
.settings-icon {
  background-color: transparent !important;
}
</style>