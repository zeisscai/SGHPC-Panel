<template>
  <div class="file-management">
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6" elevation="4">
          <v-card-title class="text-h4 font-weight-bold">
            <v-icon left class="mr-2">mdi-file-document-multiple</v-icon>
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
              :items-per-page="10"
              class="elevation-1"
            >
              <template v-slot:item.size="{ item }">
                {{ formatFileSize(item.size) }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn icon @click="downloadFile(item)">
                  <v-icon>mdi-download</v-icon>
                </v-btn>
                <v-btn icon @click="deleteFile(item)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: 'FileManagement',
  data() {
    return {
      fileHeaders: [
        { title: 'Name', key: 'name' },
        { title: 'Size', key: 'size' },
        { title: 'Modified', key: 'modified' },
        { title: 'Actions', key: 'actions', sortable: false }
      ],
      fileItems: [
        { name: 'document1.txt', size: 1024, modified: '2023-01-15' },
        { name: 'image.png', size: 204800, modified: '2023-01-10' },
        { name: 'data.csv', size: 5242880, modified: '2023-01-05' },
        { name: 'log.txt', size: 8192, modified: '2023-01-01' },
        { name: 'config.json', size: 2048, modified: '2022-12-28' }
      ]
    }
  },
  methods: {
    uploadFile() {
      // TODO: 实现文件上传功能
      console.log('Upload file')
    },
    downloadFile(file) {
      // TODO: 实现文件下载功能
      console.log('Download file', file)
    },
    deleteFile(file) {
      // TODO: 实现文件删除功能
      console.log('Delete file', file)
    },
    formatFileSize(bytes) {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
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
</style>