<template>
  <div class="page-container admin-page">
    <div class="page-header">
      <h2>用户角色管理</h2>
      <p>用于分配题目编辑、审核、重判和管理员角色。</p>
    </div>

    <div class="card">
      <el-table v-loading="loading" :data="users" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="160" />
        <el-table-column prop="email" label="邮箱" min-width="220" />
        <el-table-column prop="rating" label="Rating" width="100" />
        <el-table-column prop="registeredAt" label="注册时间" width="140" />
        <el-table-column label="角色" width="220">
          <template #default="{ row }">
            <el-select v-model="row.role" style="width: 180px" @change="value => updateRole(row, value)">
              <el-option label="普通用户" value="user" />
              <el-option label="题目编辑" value="problem_editor" />
              <el-option label="审核员" value="reviewer" />
              <el-option label="运维员" value="operator" />
              <el-option label="管理员" value="admin" />
            </el-select>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const users = ref([])
const loading = ref(false)

async function loadUsers() {
  loading.value = true
  try {
    const res = await adminApi.getUsers()
    users.value = res.data.items || []
  } finally {
    loading.value = false
  }
}

async function updateRole(row, role) {
  await adminApi.updateUserRole(row.id, { role })
  ElMessage.success(`已更新 ${row.username} 的角色`)
}

onMounted(loadUsers)
</script>

<style scoped>
.admin-page {
  max-width: 1100px;
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
</style>
