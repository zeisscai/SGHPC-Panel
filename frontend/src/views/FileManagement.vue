<template>
  <div class="file-management">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2 file-icon">mdi-file-document-multiple</v-icon>
            File Management
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
                <v-btn color="success" @click="uploadFile" class="mr-2">
                  <v-icon left>mdi-upload</v-icon>
                  Upload
                </v-btn>
                <v-btn color="info" @click="refreshFiles" :loading="loading">
                  <v-icon left>mdi-refresh</v-icon>
                  Refresh
                </v-btn>
              </div>
              <div>
                <v-text-field
                  v-model="currentPath"
                  label="Current Path"
                  dense
                  hide-details
                  class="mr-2"
                  style="width: 300px;"
                  @keyup.enter="changeDirectory"
                ></v-text-field>
                <v-btn color="primary" @click="changeDirectory">
                  Go
                </v-btn>
              </div>
            </div>
            
            <v-data-table
              :headers="fileHeaders"
              :items="fileItems"
              :loading="loading"
              :items-per-page="15"
              class="elevation-1"
              density="compact"
              item-key="name"
              show-select
              v-model="selected"
            >
              <template v-slot:item.type="{ item }">
                <v-icon v-if="item.type === 'directory'" color="blue">mdi-folder</v-icon>
                <v-icon v-else-if="isExecutable(item)" color="green">mdi-application</v-icon>
                <v-icon v-else color="grey">mdi-file</v-icon>
              </template>
              <template v-slot:item.name="{ item }">
                <span v-if="item.type === 'directory'" @click="enterDirectory(item.name)" style="cursor: pointer; color: blue; text-decoration: underline;">
                  {{ item.name }}
                </span>
                <span v-else>
                  {{ item.name }}
                </span>
              </template>
              <template v-slot:item.size="{ item }">
                {{ formatFileSize(item.size) }}
              </template>
              <template v-slot:item.permissions="{ item }">
                {{ item.permissions }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn icon @click="downloadItem(item)" class="mr-2" :disabled="item.type === 'directory'">
                  <v-icon>mdi-download</v-icon>
                </v-btn>
                <v-btn icon @click="deleteItem(item)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    
    <v-dialog v-model="uploadDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Upload File</span>
        </v-card-title>
        <v-card-text>
          <v-file-input
            v-model="fileToUpload"
            label="Select file"
            prepend-icon="mdi-paperclip"
            multiple
          ></v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="uploadDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmUpload" :loading="uploading">Upload</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <v-dialog v-model="permissionsDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Change Permissions</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newPermissions"
            label="Permissions (e.g., 0755 or rwxr-xr-x)"
            placeholder="0755"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="permissionsDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmPermissionChange" :loading="changingPermissions">Change</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import axios from 'axios'

export default {
  name: 'FileManagement',
  setup() {
    const fileHeaders = [
      { title: 'Type', key: 'type', sortable: false },
      { title: 'Name', key: 'name' },
      { title: 'Size', key: 'size' },
      { title: 'Modified', key: 'modified' },
      { title: 'Permissions', key: 'permissions' },
      { title: 'Actions', key: 'actions', sortable: false }
    ]
    
    const fileItems = ref([])
    const loading = ref(false)
    const selected = ref([])
    const currentPath = ref('.')
    
    const uploadDialog = ref(false)
    const fileToUpload = ref(null)
    const uploading = ref(false)
    
    const permissionsDialog = ref(false)
    const selectedItem = ref(null)
    const newPermissions = ref('')
    const changingPermissions = ref(false)
    
    // 页面加载时获取文件列表
    onMounted(() => {
      loadFiles()
    })
    
    // 加载文件列表
    const loadFiles = async () => {
      loading.value = true
      try {
        const response = await axios.get(`/api/file/list?path=${encodeURIComponent(currentPath.value)}`)
        fileItems.value = response.data
      } catch (error) {
        console.error('Error loading files:', error)
        alert('Error loading files: ' + (error.response?.data || error.message))
      } finally {
        loading.value = false
      }
    }
    
    // 刷新文件列表
    const refreshFiles = () => {
      loadFiles()
    }
    
    // 格式化文件大小
    const formatFileSize = (bytes) => {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }
    
    // 检查文件是否可执行
    const isExecutable = (item) => {
      return item.executable
    }
    
    // 上传文件
    const uploadFile = () => {
      uploadDialog.value = true
    }
    
    // 确认上传
    const confirmUpload = async () => {
      if (!fileToUpload.value || fileToUpload.value.length === 0) {
        alert('Please select a file to upload')
        return
      }
      
      uploading.value = true
      try {
        const formData = new FormData()
        for (let i = 0; i < fileToUpload.value.length; i++) {
          formData.append('file', fileToUpload.value[i])
        }
        formData.append('path', currentPath.value)
        
        await axios.post('/api/file/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        
        uploadDialog.value = false
        fileToUpload.value = null
        refreshFiles()
        alert('File(s) uploaded successfully')
      } catch (error) {
        console.error('Upload error:', error)
        alert('Upload failed: ' + (error.response?.data?.message || error.message))
      } finally {
        uploading.value = false
      }
    }
    
    // 下载文件
    const downloadItem = (item) => {
      if (item.type === 'directory') {
        alert('Cannot download directories')
        return
      }
      
      const fullPath = currentPath.value === '.' ? item.name : `${currentPath.value}/${item.name}`
      const downloadUrl = `/api/file/download?path=${encodeURIComponent(fullPath)}`
      window.open(downloadUrl, '_blank')
    }
    
    // 删除文件
    const deleteItem = async (item) => {
      if (!confirm(`Are you sure you want to delete ${item.name}?`)) {
        return
      }
      
      try {
        const fullPath = currentPath.value === '.' ? item.name : `${currentPath.value}/${item.name}`
        await axios.delete(`/api/file/delete?path=${encodeURIComponent(fullPath)}`)
        refreshFiles()
        alert('File deleted successfully')
      } catch (error) {
        console.error('Delete error:', error)
        alert('Delete failed: ' + (error.response?.data?.message || error.message))
      }
    }
    
    // 进入目录
    const enterDirectory = (dirName) => {
      if (currentPath.value === '.') {
        currentPath.value = dirName
      } else {
        currentPath.value = `${currentPath.value}/${dirName}`
      }
      loadFiles()
    }
    
    // 更改目录
    const changeDirectory = () => {
      loadFiles()
    }
    
    // 打开权限更改对话框
    const openPermissionsDialog = (item) => {
      selectedItem.value = item
      newPermissions.value = item.permissions
      permissionsDialog.value = true
    }
    
    // 确认权限更改
    const confirmPermissionChange = async () => {
      if (!selectedItem.value) return
      
      changingPermissions.value = true
      try {
        const fullPath = currentPath.value === '.' ? selectedItem.value.name : `${currentPath.value}/${selectedItem.value.name}`
        await axios.put('/api/file/permissions', {
          path: fullPath,
          permissions: newPermissions.value
        })
        
        permissionsDialog.value = false
        selectedItem.value = null
        newPermissions.value = ''
        refreshFiles()
        alert('Permissions changed successfully')
      } catch (error) {
        console.error('Permission change error:', error)
        alert('Permission change failed: ' + (error.response?.data?.message || error.message))
      } finally {
        changingPermissions.value = false
      }
    }
    
    return {
      fileHeaders,
      fileItems,
      loading,
      selected,
      currentPath,
      uploadDialog,
      fileToUpload,
      uploading,
      permissionsDialog,
      newPermissions,
      changingPermissions,
      formatFileSize,
      isExecutable,
      uploadFile,
      confirmUpload,
      downloadItem,
      deleteItem,
      enterDirectory,
      refreshFiles,
      changeDirectory,
      openPermissionsDialog,
      confirmPermissionChange
    }
  }
}
</script>

<style scoped>
.file-management {
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

.file-icon {
  background-color: transparent !important;
}
</style>