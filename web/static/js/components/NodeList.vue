<template>
  <div class="card">
    <h2>节点列表</h2>
    <table v-if="nodes.length > 0">
      <thead>
        <tr>
          <th>类型</th>
          <th>IP地址</th>
          <th>主机名</th>
          <th>状态</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(node, index) in nodes" :key="index">
          <td>{{ node.type === 'master' ? 'Master' : 'Compute' }}</td>
          <td>{{ node.ip }}</td>
          <td>{{ node.hostname }}</td>
          <td>{{ node.status || '未部署' }}</td>
          <td>
            <button @click="removeNode(index)" class="btn btn-danger">删除</button>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else>暂无节点配置</p>
  </div>
</template>

<script>
export default {
  name: 'NodeList',
  props: {
    nodes: {
      type: Array,
      default: () => []
    }
  },
  methods: {
    removeNode(index) {
      this.$emit('remove-node', index);
    }
  }
}
</script>