<template>
  <div class="page-container admin-page">
    <div class="page-header">
      <h2>审计日志</h2>
      <p>查看题目、重判等后台操作记录。</p>
    </div>

    <div class="card filter-bar">
      <el-input v-model="filters.username" placeholder="按用户名筛选" clearable style="width: 180px" />
      <el-input v-model="filters.resourceType" placeholder="资源类型" clearable style="width: 180px" />
      <el-input v-model="filters.action" placeholder="操作类型" clearable style="width: 180px" />
      <el-button type="primary" @click="loadLogs">查询</el-button>
    </div>

    <div class="card">
      <el-table v-loading="loading" :data="logs" stripe>
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="username" label="用户" width="140" />
        <el-table-column prop="userRole" label="角色" width="140" />
        <el-table-column prop="resourceType" label="资源类型" width="140" />
        <el-table-column prop="resourceId" label="资源 ID" width="120" />
        <el-table-column prop="action" label="操作" width="120" />
        <el-table-column prop="detail" label="详情" min-width="260" />
        <el-table-column prop="createdAt" label="时间" width="200" />
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const logs = ref([])
const filters = reactive({
  username: '',
  resourceType: '',
  action: ''
})

async function loadLogs() {
  loading.value = true
  try {
    const res = await adminApi.getAuditLogs(filters)
    logs.value = res.data.items || []
  } finally {
    loading.value = false
  }
}

onMounted(loadLogs)
</script>

<style scoped>
.admin-page {
  max-width: 1200px;
}
.page-header {
  margin-bottom: 20px;
}
.page-header h2 {
  font-size: 28px;
  margin-bottom: 6px;
}
.page-header p {
  color: var(--text-secondary);
}
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
</style>
