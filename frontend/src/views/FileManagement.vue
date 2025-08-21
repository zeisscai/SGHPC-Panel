<template>
  <div class="file-management" @dragover.prevent="onDragOver" @drop.prevent="onDrop" @dragleave.prevent="onDragLeave">
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
                <v-btn color="info" @click="refreshFiles" :loading="loading" class="mr-2">
                  <v-icon left>mdi-refresh</v-icon>
                  Refresh
                </v-btn>
                <v-btn color="warning" @click="createNewFolder" class="mr-2">
                  <v-icon left>mdi-folder-plus</v-icon>
                  New Folder
                </v-btn>
                <v-btn color="error" @click="deleteSelected" :disabled="selected.length === 0">
                  <v-icon left>mdi-delete-sweep</v-icon>
                  Delete Selected
                </v-btn>
              </div>
              <div class="d-flex align-center">
                <v-btn icon @click="navigateUp" class="mr-2" :disabled="currentPath === '.'">
                  <v-icon>mdi-arrow-up</v-icon>
                </v-btn>
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
            
            <div class="breadcrumbs mb-2" v-if="currentPath !== '.'">
              <v-breadcrumbs :items="pathBreadcrumbs" divider="/">
                <template v-slot:item="{ item }">
                  <v-breadcrumbs-item
                    :text="item.text"
                    @click="navigateToBreadcrumb(item.path)"
                  ></v-breadcrumbs-item>
                </template>
              </v-breadcrumbs>
            </div>
            
            <div class="drop-zone" v-if="isDragging">
              <div class="drop-zone-content">
                <v-icon size="64" color="primary">mdi-cloud-upload</v-icon>
                <h3>Drop files here to upload</h3>
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
                <v-btn icon x-small @click="openPermissionsDialog(item)" class="ml-2">
                  <v-icon small>mdi-pencil</v-icon>
                </v-btn>
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn icon @click="downloadItem(item)" class="mr-2" :disabled="item.type === 'directory'">
                  <v-icon>mdi-download</v-icon>
                </v-btn>
                <v-btn icon @click="deleteItem(item)" class="mr-2">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
                <v-btn icon @click="renameItem(item)">
                  <v-icon>mdi-rename</v-icon>
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
            show-size
            counter
            chips
            truncate-length="15"
          ></v-file-input>
          <v-alert
            v-if="isDragging"
            type="info"
            class="mt-3"
            dense
          >
            You can also drag and drop files anywhere on this page
          </v-alert>
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
    
    <v-dialog v-model="newFolderDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Create New Folder</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newFolderName"
            label="Folder Name"
            placeholder="New Folder"
            @keyup.enter="confirmCreateFolder"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="newFolderDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmCreateFolder" :loading="creatingFolder">Create</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <v-dialog v-model="renameDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Rename Item</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newItemName"
            label="New Name"
            @keyup.enter="confirmRename"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="renameDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmRename" :loading="renaming">Rename</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    

  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
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
    const isDragging = ref(false)
    
    // 上传相关
    const uploadDialog = ref(false)
    const fileToUpload = ref(null)
    const uploading = ref(false)
    
    // 权限相关
    const permissionsDialog = ref(false)
    const selectedItem = ref(null)
    const newPermissions = ref('')
    const changingPermissions = ref(false)
    
    // 新文件夹相关
    const newFolderDialog = ref(false)
    const newFolderName = ref('')
    const creatingFolder = ref(false)
    
    // 重命名相关
    const renameDialog = ref(false)
    const newItemName = ref('')
    const renaming = ref(false)
    

    
    // 计算面包屑导航
    const pathBreadcrumbs = computed(() => {
      if (currentPath.value === '.') return []
      
      const parts = currentPath.value.split('/')
      let path = ''
      
      return [{ text: 'Root', path: '.' }].concat(
        parts.map(part => {
          path = path ? `${path}/${part}` : part
          return { text: part, path }
        })
      )
    })
    
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
    

    
    // 拖放相关函数
    const onDragOver = (event) => {
      isDragging.value = true
    }
    
    const onDragLeave = (event) => {
      isDragging.value = false
    }
    
    const onDrop = (event) => {
      isDragging.value = false
      const files = event.dataTransfer.files
      if (files.length > 0) {
        fileToUpload.value = files
        confirmUpload()
      }
    }
    
    // 上传文件
    const uploadFile = () => {
      uploadDialog.value = true
    }
    
    // 确认上传
    const confirmUpload = async () => {
      if (!fileToUpload.value || fileToUpload.value.length === 0) {
        alert('请选择要上传的文件')
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
          },
          onUploadProgress: (progressEvent) => {
            // 可以在这里添加上传进度处理
            const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
            console.log(`上传进度: ${percentCompleted}%`)
          }
        })
        
        uploadDialog.value = false
        fileToUpload.value = null
        refreshFiles()
        alert('文件上传成功')
      } catch (error) {
        console.error('上传错误:', error)
        alert('上传失败: ' + (error.response?.data?.message || error.message))
      } finally {
        uploading.value = false
      }
    }
    
    // 获取文件完整路径
    const getFullPath = (item) => {
      return currentPath.value === '.' ? item.name : `${currentPath.value}/${item.name}`
    }
    
    // 下载文件
    const downloadItem = (item) => {
      if (item.type === 'directory') {
        alert('不能下载目录')
        return
      }
      
      const fullPath = getFullPath(item)
      const downloadUrl = `/api/file/download?path=${encodeURIComponent(fullPath)}`
      window.open(downloadUrl, '_blank')
    }
    
    // 删除文件
    const deleteItem = async (item) => {
      if (!confirm(`确定要删除 ${item.name}?`)) {
        return
      }
      
      try {
        const fullPath = getFullPath(item)
        await axios.delete(`/api/file/delete?path=${encodeURIComponent(fullPath)}`)
        refreshFiles()
        alert('文件删除成功')
      } catch (error) {
        console.error('删除错误:', error)
        alert('删除失败: ' + (error.response?.data?.message || error.message))
      }
    }
    
    // 删除选中的多个文件
    const deleteSelected = async () => {
      if (selected.value.length === 0) return
      
      if (!confirm(`确定要删除选中的 ${selected.value.length} 个项目?`)) {
        return
      }
      
      let successCount = 0
      let failCount = 0
      
      for (const item of selected.value) {
        try {
          const fullPath = getFullPath(item)
          await axios.delete(`/api/file/delete?path=${encodeURIComponent(fullPath)}`)
          successCount++
        } catch (error) {
          console.error(`删除 ${item.name} 失败:`, error)
          failCount++
        }
      }
      
      refreshFiles()
      alert(`删除完成: ${successCount} 个成功, ${failCount} 个失败`)
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
    
    // 导航到上一级目录
    const navigateUp = () => {
      if (currentPath.value === '.') return
      
      const parts = currentPath.value.split('/')
      parts.pop()
      currentPath.value = parts.length === 0 ? '.' : parts.join('/')
      loadFiles()
    }
    
    // 导航到面包屑指定路径
    const navigateToBreadcrumb = (path) => {
      currentPath.value = path
      loadFiles()
    }
    
    // 更改目录
    const changeDirectory = () => {
      loadFiles()
    }
    
    // 创建新文件夹
    const createNewFolder = () => {
      newFolderName.value = ''
      newFolderDialog.value = true
    }
    
    // 确认创建文件夹
    const confirmCreateFolder = async () => {
      if (!newFolderName.value) {
        alert('请输入文件夹名称')
        return
      }
      
      creatingFolder.value = true
      try {
        const fullPath = currentPath.value === '.' ? 
          newFolderName.value : 
          `${currentPath.value}/${newFolderName.value}`
        
        await axios.post('/api/file/mkdir', {
          path: fullPath
        })
        
        newFolderDialog.value = false
        refreshFiles()
        alert('文件夹创建成功')
      } catch (error) {
        console.error('创建文件夹错误:', error)
        alert('创建文件夹失败: ' + (error.response?.data?.message || error.message))
      } finally {
        creatingFolder.value = false
      }
    }
    
    // 重命名文件或目录
    const renameItem = (item) => {
      selectedItem.value = item
      newItemName.value = item.name
      renameDialog.value = true
    }
    
    // 确认重命名
    const confirmRename = async () => {
      if (!selectedItem.value || !newItemName.value) return
      if (newItemName.value === selectedItem.value.name) {
        renameDialog.value = false
        return
      }
      
      renaming.value = true
      try {
        const oldPath = getFullPath(selectedItem.value)
        const newPath = currentPath.value === '.' ? 
          newItemName.value : 
          `${currentPath.value}/${newItemName.value}`
        
        await axios.post('/api/file/rename', {
          oldPath,
          newPath
        })
        
        renameDialog.value = false
        refreshFiles()
        alert('重命名成功')
      } catch (error) {
        console.error('重命名错误:', error)
        alert('重命名失败: ' + (error.response?.data?.message || error.message))
      } finally {
        renaming.value = false
      }
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
        const fullPath = getFullPath(selectedItem.value)
        await axios.put('/api/file/permissions', {
          path: fullPath,
          permissions: newPermissions.value
        })
        
        permissionsDialog.value = false
        selectedItem.value = null
        newPermissions.value = ''
        refreshFiles()
        alert('权限修改成功')
      } catch (error) {
        console.error('权限修改错误:', error)
        alert('权限修改失败: ' + (error.response?.data?.message || error.message))
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