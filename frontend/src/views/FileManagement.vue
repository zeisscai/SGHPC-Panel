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
              <v-btn color="success" @click="uploadFile" class="mr-2">
                <v-icon left>mdi-upload</v-icon>
                Upload
              </v-btn>
              <v-btn color="info" @click="downloadFile">
                <v-icon left>mdi-download</v-icon>
                Download
              </v-btn>
            </div>
            
            <v-data-table
              :headers="fileHeaders"
              :items="fileItems"
              :loading="loading"
              :items-per-page="10"
              class="elevation-1"
              density="compact"
              item-key="name"
              show-select
              v-model="selected"
            >
              <template v-slot:item.type="{ item }">
                <v-icon v-if="item.type === 'folder'" color="blue">mdi-folder</v-icon>
                <v-icon v-else color="grey">mdi-file</v-icon>
              </template>
              <template v-slot:item.size="{ item }">
                {{ formatFileSize(item.size) }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn icon @click="downloadItem(item)" class="mr-2">
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
          ></v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="uploadDialog = false">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="confirmUpload">Upload</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { ref } from 'vue'
import axios from 'axios'

export default {
  name: 'FileManagement',
  setup() {
    const fileHeaders = [
      { title: 'Type', key: 'type', sortable: false },
      { title: 'Name', key: 'name' },
      { title: 'Size', key: 'size' },
      { title: 'Modified', key: 'modified' },
      { title: 'Actions', key: 'actions', sortable: false }
    ]
    
    const fileItems = ref([
      { type: 'folder', name: 'Documents', size: 0, modified: '2023-01-15 14:30' },
      { type: 'folder', name: 'Projects', size: 0, modified: '2023-01-10 09:15' },
      { type: 'file', name: 'readme.txt', size: 1024, modified: '2023-01-05 16:45' },
      { type: 'file', name: 'config.json', size: 2048, modified: '2023-01-03 11:20' }
    ])
    
    const loading = ref(false)
    const selected = ref([])
    
    const uploadDialog = ref(false)
    const fileToUpload = ref(null)
    
    const formatFileSize = (bytes) => {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }
    
    const uploadFile = () => {
      uploadDialog.value = true
    }
    
    const confirmUpload = () => {
      if (fileToUpload.value) {
        // 模拟上传过程
        loading.value = true
        setTimeout(() => {
          loading.value = false
          uploadDialog.value = false
          fileToUpload.value = null
          // 添加新文件到列表顶部
          fileItems.value.unshift({
            type: 'file',
            name: 'new-file.txt',
            size: Math.floor(Math.random() * 10000),
            modified: new Date().toLocaleString()
          })
        }, 1500)
      }
    }
    
    const downloadFile = () => {
      alert('Download functionality would be implemented here')
    }
    
    const downloadItem = (item) => {
      alert(`Downloading ${item.name}`)
    }
    
    const deleteItem = (item) => {
      const index = fileItems.value.indexOf(item)
      if (index !== -1) {
        fileItems.value.splice(index, 1)
      }
    }
    
    return {
      fileHeaders,
      fileItems,
      loading,
      selected,
      uploadDialog,
      fileToUpload,
      formatFileSize,
      uploadFile,
      confirmUpload,
      downloadFile,
      downloadItem,
      deleteItem
    }
  }
}
</script>

<style scoped>
.file-icon {
  background-color: transparent !important;
}
</style>