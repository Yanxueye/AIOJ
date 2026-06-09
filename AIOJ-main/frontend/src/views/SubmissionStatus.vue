<template>
  <div class="status-page page-container">
    <div class="page-header">
      <h2>评测状态</h2>
    </div>

    <div class="filter-bar card">
      <el-input
        v-model="filters.problemId"
        placeholder="题号"
        clearable
        style="width: 120px"
        @input="debouncedLoad"
      />
      <el-select v-model="filters.status" placeholder="评测状态" clearable style="width: 220px" @change="loadSubmissions">
        <el-option v-for="s in statusOptions" :key="s" :label="s" :value="s" />
      </el-select>
      <el-select v-model="filters.sortBy" style="width: 140px" @change="loadSubmissions">
        <el-option label="按时间排序" value="time" />
        <el-option label="按题号排序" value="problemId" />
      </el-select>
      <div style="flex: 1" />
      <el-button @click="loadSubmissions">
        <el-icon><Refresh /></el-icon>刷新
      </el-button>
    </div>

    <div class="card">
      <el-table
        :data="submissionStore.submissions"
        v-loading="submissionStore.loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="提交编号" width="110" />
        <el-table-column label="题号" width="80">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.problemId}`" class="link">
              {{ row.problemId }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column prop="problemTitle" label="题目名称" min-width="180">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.problemId}`" class="link">
              {{ row.problemTitle }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column label="评测结果" width="200">
          <template #default="{ row }">
            <span :class="statusClass(row.status)">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column label="运行时间" width="110">
          <template #default="{ row }">
            {{ row.runtimeMs != null ? row.runtimeMs + 'ms' : (row.runtime != null ? row.runtime + 'ms' : '-') }}
          </template>
        </el-table-column>
        <el-table-column label="内存" width="150">
          <template #default="{ row }">
            <span v-if="row.memoryKb != null && row.memoryKb > 0">{{ row.memoryKb }} KB</span>
            <span v-else-if="row.memory != null">{{ row.memory }} MB</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="submissionStore.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadSubmissions"
          @size-change="loadSubmissions"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { useSubmissionStore } from '@/stores/submission'

const submissionStore = useSubmissionStore()

const statusOptions = [
  'Pending',
  'Queueing',
  'Compiling',
  'Running',
  'Accepted',
  'Wrong Answer',
  'Compile Error',
  'Runtime Error',
  'Time Limit Exceeded',
  'Memory Limit Exceeded',
  'Output Limit Exceeded',
  'System Error'
]

const filters = reactive({ problemId: '', status: '', sortBy: 'time' })
const pagination = reactive({ page: 1, pageSize: 20 })

let loadTimer = null
function debouncedLoad() {
  clearTimeout(loadTimer)
  loadTimer = setTimeout(loadSubmissions, 300)
}

function loadSubmissions() {
  submissionStore.fetchSubmissions({
    page: pagination.page,
    pageSize: pagination.pageSize,
    problemId: filters.problemId,
    status: filters.status,
    sortBy: filters.sortBy
  })
}

function statusClass(status) {
  const map = {
    'Pending': 'status-pending',
    'Queueing': 'status-pending',
    'Compiling': 'status-running',
    'Running': 'status-running',
    'Accepted': 'status-accepted',
    'Wrong Answer': 'status-wrong',
    'Compile Error': 'status-ce',
    'Runtime Error': 'status-wrong',
    'Time Limit Exceeded': 'status-tle',
    'Memory Limit Exceeded': 'status-mle',
    'Output Limit Exceeded': 'status-ole',
    'System Error': 'status-system'
  }
  return map[status] || ''
}

function formatTime(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

onMounted(loadSubmissions)
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}
.page-header h2 {
  font-size: 24px;
  font-weight: 700;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.link {
  color: var(--accent-blue);
  font-weight: 500;
}
.link:hover {
  text-decoration: underline;
}
.pagination-wrap {
  display: flex;
  justify-content: center;
  padding-top: 20px;
}
</style>
