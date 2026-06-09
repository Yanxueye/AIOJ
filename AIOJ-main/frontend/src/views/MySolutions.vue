<template>
  <div class="page-container solutions-page">
    <div class="page-header">
      <h2>我的题解</h2>
      <p>管理你保存的草稿和已发布题解。</p>
    </div>

    <div class="card">
      <el-table v-loading="loading" :data="solutions" stripe>
        <el-table-column prop="problemId" label="题号" width="100" />
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column prop="language" label="语言" width="120" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.isPublished ? 'success' : 'info'" size="small">
              {{ row.isPublished ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <router-link :to="`/my/solutions/${row.id}/edit`" class="solution-link">编辑</router-link>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { problemApi } from '@/api/problem'

const loading = ref(true)
const solutions = ref([])

onMounted(async () => {
  try {
    const res = await problemApi.getMySolutions()
    solutions.value = res.data.items || []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
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
.solution-link {
  color: var(--accent-blue);
  font-weight: 600;
}
</style>
