<template>
  <v-card class="table-component" elevation="2">
    <v-data-table
      :headers="headers"
      :items="data"
      :items-per-page="10"
      class="elevation-1"
    >
      <template v-slot:item="{ item }">
        <tr>
          <td v-for="column in columns" :key="column.key">
            {{ formatCellData(item[column.key], column.key) }}
          </td>
        </tr>
      </template>
    </v-data-table>
  </v-card>
</template>

<script>
export default {
  name: 'TableComponent',
  props: {
    data: {
      type: Array,
      required: true
    },
    columns: {
      type: Array,
      required: true
    }
  },
  computed: {
    headers() {
      return this.columns.map(column => ({
        title: column.label,
        key: column.key
      }))
    }
  },
  methods: {
    formatCellData(value, key) {
      if (value === null || value === undefined) {
        return '-'
      }
      
      // 格式化CPU和内存使用率
      if (key === 'cpu_usage' || key === 'memory_usage') {
        return typeof value === 'number' ? value.toFixed(2) + '%' : value
      }
      
      // 格式化时间
      if (key === 'submission_time') {
        return new Date(value).toLocaleString()
      }
      
      return value
    }
  }
}
</script>

<style scoped>
.table-component {
  border-radius: 8px;
  overflow: hidden;
  transition: box-shadow 0.3s ease;
}

.table-component:hover {
  box-shadow: 0 8px 16px rgba(0,0,0,0.2);
}
</style>