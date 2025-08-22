<template>
  <div class="user-management">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2 user-icon">mdi-account-multiple</v-icon>
            User
          </v-card-title>
        </v-card>
      </v-col>
    </v-row>
    
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="2" transition="slide-y-reverse-transition">
          <v-card-text>
            <div class="d-flex justify-space-between mb-4">
              <div>
                <v-btn color="success" @click="addUser" class="mr-2">
                  <v-icon left>mdi-account-plus</v-icon>
                  Add User
                </v-btn>
                <v-btn color="info" @click="refreshUsers" :loading="loading" class="mr-2">
                  <v-icon left>mdi-refresh</v-icon>
                  Refresh
                </v-btn>
                <v-btn color="warning" @click="generateKeysForSelected" :disabled="selected.length === 0" class="mr-2">
                  <v-icon left>mdi-key</v-icon>
                  Generate Keys
                </v-btn>
                <v-btn color="error" @click="deleteSelected" :disabled="selected.length === 0">
                  <v-icon left>mdi-delete-sweep</v-icon>
                  Delete Selected
                </v-btn>
              </div>
            </div>
            
            <v-data-table
              :headers="userHeaders"
              :items="userItems"
              :loading="loading"
              :items-per-page="15"
              class="elevation-1"
              density="compact"
              item-key="username"
              show-select
              v-model="selected"
            >
              <template v-slot:item.sudo="{ item }">
                <v-chip
                  :color="item.sudo ? 'success' : 'default'"
                  :text-color="item.sudo ? 'white' : 'black'"
                  small
                >
                  {{ item.sudo ? 'Yes' : 'No' }}
                </v-chip>
              </template>
              <template v-slot:item.status="{ item }">
                <v-chip
                  :color="item.status === 'active' ? 'success' : 'error'"
                  :text-color="'white'"
                  small
                >
                  {{ item.status === 'active' ? 'Active' : 'Disabled' }}
                </v-chip>
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn icon @click="changePassword(item)" class="mr-1" size="small">
                  <v-icon small>mdi-lock-reset</v-icon>
                </v-btn>
                <v-btn icon @click="generateKey(item)" class="mr-1" size="small">
                  <v-icon small>mdi-key</v-icon>
                </v-btn>
                <v-btn 
                  icon 
                  @click="toggleUserStatus(item)" 
                  class="mr-1" 
                  size="small"
                  :color="item.status === 'active' ? 'warning' : 'success'"
                >
                  <v-icon small>{{ item.status === 'active' ? 'mdi-account-off' : 'mdi-account-check' }}</v-icon>
                </v-btn>
                <v-btn icon @click="deleteUser(item)" size="small" color="error">
                  <v-icon small>mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <!-- Add User Dialog -->
    <v-dialog v-model="addUserDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Add New User</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newUser.username"
            label="Username"
            required
            :rules="[rules.required]"
          ></v-text-field>
          <v-text-field
            v-model="newUser.password"
            label="Password"
            type="password"
            required
            :rules="[rules.required]"
          ></v-text-field>
          <v-text-field
            v-model="newUser.groups"
            label="Groups (comma separated)"
            placeholder="users,developers"
          ></v-text-field>
          <v-checkbox
            v-model="newUser.sudo"
            label="Grant sudo privileges"
          ></v-checkbox>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="addUserDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmAddUser" :loading="addingUser">Add User</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- Change Password Dialog -->
    <v-dialog v-model="passwordDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Change Password for {{ selectedUser?.username }}</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newPassword"
            label="New Password"
            type="password"
            required
            :rules="[rules.required]"
          ></v-text-field>
          <v-text-field
            v-model="confirmPassword"
            label="Confirm Password"
            type="password"
            required
            :rules="[rules.required, rules.passwordMatch]"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="passwordDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmPasswordChange" :loading="changingPassword">Change Password</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- Generate Key Dialog -->
    <v-dialog v-model="keyDialog" max-width="600px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Generate SSH Key for {{ selectedUser?.username }}</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="keyType"
            :items="keyTypes"
            label="Key Type"
            required
          ></v-select>
          <v-text-field
            v-model="keyComment"
            label="Comment (optional)"
            placeholder="user@hostname"
          ></v-text-field>
          <v-alert v-if="generatedKey" type="success" class="mt-3">
            <strong>Key generated successfully!</strong><br>
            <small>The private key has been saved to the user's home directory.</small>
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="keyDialog = false">Close</v-btn>
          <v-btn color="blue darken-1" text @click="confirmGenerateKey" :loading="generatingKey">Generate Key</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

// 响应式数据
const loading = ref(false)
const userItems = ref([])
const selected = ref([])
const addUserDialog = ref(false)
const passwordDialog = ref(false)
const keyDialog = ref(false)
const addingUser = ref(false)
const changingPassword = ref(false)
const generatingKey = ref(false)
const selectedUser = ref(null)
const generatedKey = ref(false)

// 表单数据
const newUser = ref({
  username: '',
  password: '',
  groups: '',
  sudo: false
})

const newPassword = ref('')
const confirmPassword = ref('')
const keyType = ref('rsa')
const keyComment = ref('')

// 表格头部定义
const userHeaders = [
  { title: 'Username', key: 'username', sortable: true },
  { title: 'Groups', key: 'groups', sortable: false },
  { title: 'Sudo', key: 'sudo', sortable: true },
  { title: 'Status', key: 'status', sortable: true },
  { title: 'Last Login', key: 'lastLogin', sortable: true },
  { title: 'Actions', key: 'actions', sortable: false, width: '200px' }
]

// SSH密钥类型
const keyTypes = [
  { title: 'RSA (2048 bit)', value: 'rsa' },
  { title: 'RSA (4096 bit)', value: 'rsa4096' },
  { title: 'ED25519', value: 'ed25519' },
  { title: 'ECDSA', value: 'ecdsa' }
]

// 表单验证规则
const rules = {
  required: value => !!value || 'This field is required',
  passwordMatch: value => value === newPassword.value || 'Passwords do not match'
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/users')
    userItems.value = response.data.users || []
  } catch (error) {
    console.error('加载用户列表失败:', error)
    // 模拟数据用于演示
    userItems.value = [
      {
        username: 'admin',
        groups: 'wheel,sudo',
        sudo: true,
        status: 'active',
        lastLogin: '2024-01-15 10:30:00'
      },
      {
        username: 'user1',
        groups: 'users',
        sudo: false,
        status: 'active',
        lastLogin: '2024-01-14 15:20:00'
      },
      {
        username: 'testuser',
        groups: 'users,developers',
        sudo: false,
        status: 'disabled',
        lastLogin: '2024-01-10 09:15:00'
      }
    ]
  } finally {
    loading.value = false
  }
}

// 刷新用户列表
const refreshUsers = () => {
  loadUsers()
}

// 添加用户
const addUser = () => {
  newUser.value = {
    username: '',
    password: '',
    groups: '',
    sudo: false
  }
  addUserDialog.value = true
}

// 确认添加用户
const confirmAddUser = async () => {
  if (!newUser.value.username || !newUser.value.password) {
    alert('请填写用户名和密码')
    return
  }
  
  addingUser.value = true
  try {
    await axios.post('/api/users', newUser.value)
    addUserDialog.value = false
    refreshUsers()
    alert('用户添加成功')
  } catch (error) {
    console.error('添加用户失败:', error)
    alert('添加用户失败: ' + (error.response?.data?.message || error.message))
  } finally {
    addingUser.value = false
  }
}

// 修改密码
const changePassword = (user) => {
  selectedUser.value = user
  newPassword.value = ''
  confirmPassword.value = ''
  passwordDialog.value = true
}

// 确认修改密码
const confirmPasswordChange = async () => {
  if (!newPassword.value || newPassword.value !== confirmPassword.value) {
    alert('请确保密码填写正确且两次输入一致')
    return
  }
  
  if (!confirm(`确定要修改用户 ${selectedUser.value.username} 的密码吗？`)) {
    return
  }
  
  changingPassword.value = true
  try {
    await axios.put(`/api/users/${selectedUser.value.username}/password`, {
      password: newPassword.value
    })
    passwordDialog.value = false
    alert('密码修改成功')
  } catch (error) {
    console.error('修改密码失败:', error)
    alert('修改密码失败: ' + (error.response?.data?.message || error.message))
  } finally {
    changingPassword.value = false
  }
}

// 生成密钥
const generateKey = (user) => {
  selectedUser.value = user
  keyType.value = 'rsa'
  keyComment.value = `${user.username}@hostname`
  generatedKey.value = false
  keyDialog.value = true
}

// 确认生成密钥
const confirmGenerateKey = async () => {
  generatingKey.value = true
  try {
    await axios.post(`/api/users/${selectedUser.value.username}/generate-key`, {
      type: keyType.value,
      comment: keyComment.value
    })
    generatedKey.value = true
    alert('SSH密钥生成成功')
  } catch (error) {
    console.error('生成密钥失败:', error)
    alert('生成密钥失败: ' + (error.response?.data?.message || error.message))
  } finally {
    generatingKey.value = false
  }
}

// 切换用户状态
const toggleUserStatus = async (user) => {
  const action = user.status === 'active' ? '停用' : '启用'
  if (!confirm(`确定要${action}用户 ${user.username} 吗？`)) {
    return
  }
  
  try {
    const newStatus = user.status === 'active' ? 'disabled' : 'active'
    await axios.put(`/api/users/${user.username}/status`, {
      status: newStatus
    })
    user.status = newStatus
    alert(`用户${action}成功`)
  } catch (error) {
    console.error(`${action}用户失败:`, error)
    alert(`${action}用户失败: ` + (error.response?.data?.message || error.message))
  }
}

// 删除用户
const deleteUser = async (user) => {
  if (!confirm(`确定要删除用户 ${user.username} 吗？此操作不可撤销！`)) {
    return
  }
  
  try {
    await axios.delete(`/api/users/${user.username}`)
    refreshUsers()
    alert('用户删除成功')
  } catch (error) {
    console.error('删除用户失败:', error)
    alert('删除用户失败: ' + (error.response?.data?.message || error.message))
  }
}

// 删除选中的用户
const deleteSelected = async () => {
  if (selected.value.length === 0) return
  
  if (!confirm(`确定要删除选中的 ${selected.value.length} 个用户吗？此操作不可撤销！`)) {
    return
  }
  
  let successCount = 0
  let failCount = 0
  
  for (const user of selected.value) {
    try {
      await axios.delete(`/api/users/${user.username}`)
      successCount++
    } catch (error) {
      console.error(`删除用户 ${user.username} 失败:`, error)
      failCount++
    }
  }
  
  refreshUsers()
  selected.value = []
  alert(`删除完成: ${successCount} 个成功, ${failCount} 个失败`)
}

// 为选中用户生成密钥
const generateKeysForSelected = async () => {
  if (selected.value.length === 0) return
  
  if (!confirm(`确定要为选中的 ${selected.value.length} 个用户生成SSH密钥吗？`)) {
    return
  }
  
  let successCount = 0
  let failCount = 0
  
  for (const user of selected.value) {
    try {
      await axios.post(`/api/users/${user.username}/generate-key`, {
        type: 'rsa',
        comment: `${user.username}@hostname`
      })
      successCount++
    } catch (error) {
      console.error(`为用户 ${user.username} 生成密钥失败:`, error)
      failCount++
    }
  }
  
  selected.value = []
  alert(`密钥生成完成: ${successCount} 个成功, ${failCount} 个失败`)
}

// 组件挂载时加载数据
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.user-icon {
  color: #1976d2;
  background-color: transparent !important;
}

.v-card {
  border-radius: 8px;
}

.v-btn {
  text-transform: none;
}

.v-chip {
  font-weight: 500;
}

.v-icon {
  background-color: transparent !important;
}
</style>