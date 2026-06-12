<template>
  <div class="status-page page-container">
    <div class="page-header">
      <h2>评测状态</h2>
      <p class="page-desc">查看提交记录和评测结果</p>
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
        <el-table-column label="提交编号" width="110">
          <template #default="{ row }">
            <span class="sub-id-link" @click="openCodeDialog(row)">{{ row.id }}</span>
          </template>
        </el-table-column>
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

    <!-- Code Dialog -->
    <el-dialog
      v-model="codeDialogVisible"
      :title="`提交 #${codeDialogData.id} — ${codeDialogData.problemTitle}`"
      width="800px"
      top="5vh"
      destroy-on-close
    >
      <div class="code-dialog-content">
        <div class="code-dialog-meta">
          <span :class="statusClass(codeDialogData.status)">{{ codeDialogData.status }}</span>
          <span class="meta-sep">·</span>
          <span>{{ codeDialogData.language }}</span>
          <span class="meta-sep">·</span>
          <span>{{ codeDialogData.runtimeMs != null ? codeDialogData.runtimeMs + 'ms' : '-' }}</span>
          <span class="meta-sep">·</span>
          <span>{{ codeDialogData.memoryKb != null ? codeDialogData.memoryKb + ' KB' : '-' }}</span>
          <span class="meta-sep">·</span>
          <span>{{ formatTime(codeDialogData.createdAt) }}</span>
        </div>
        <div v-if="codeDialogData.errorMessage" class="code-dialog-error">
          <div class="error-label">错误信息</div>
          <pre class="error-pre">{{ codeDialogData.errorMessage }}</pre>
        </div>
        <div v-if="codeDialogData.compileOutput" class="code-dialog-error">
          <div class="error-label">编译输出</div>
          <pre class="error-pre">{{ codeDialogData.compileOutput }}</pre>
        </div>
        <div class="code-dialog-code">
          <div class="code-label">源代码</div>
          <div v-if="loadingCodeDialog" class="code-loading">
            <el-icon class="is-loading"><Loading /></el-icon> 加载中...
          </div>
          <pre v-else class="code-block"><code>{{ codeDialogCode }}</code></pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useSubmissionStore } from '@/stores/submission'
import { submissionApi } from '@/api/submission'

const submissionStore = useSubmissionStore()

const statusOptions = [
  'Pending', 'Queueing', 'Compiling', 'Running', 'Accepted',
  'Wrong Answer', 'Compile Error', 'Runtime Error',
  'Time Limit Exceeded', 'Memory Limit Exceeded', 'Output Limit Exceeded', 'System Error'
]

const filters = reactive({ problemId: '', status: '', sortBy: 'time' })
const pagination = reactive({ page: 1, pageSize: 20 })

// Code dialog state
const codeDialogVisible = ref(false)
const codeDialogData = ref({})
const codeDialogCode = ref('')
const loadingCodeDialog = ref(false)

async function openCodeDialog(row) {
  codeDialogData.value = row
  codeDialogVisible.value = true
  codeDialogCode.value = ''
  loadingCodeDialog.value = true
  try {
    const res = await submissionApi.getDetail(row.id)
    codeDialogCode.value = res.data?.code || '暂无代码'
    // Update meta with full detail
    if (res.data) {
      codeDialogData.value = { ...row, ...res.data }
    }
  } catch {
    codeDialogCode.value = '加载失败'
  } finally {
    loadingCodeDialog.value = false
  }
}

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
    'Pending': 'status-pending', 'Queueing': 'status-pending',
    'Compiling': 'status-running', 'Running': 'status-running',
    'Accepted': 'status-accepted', 'Wrong Answer': 'status-wrong',
    'Compile Error': 'status-ce', 'Runtime Error': 'status-wrong',
    'Time Limit Exceeded': 'status-tle', 'Memory Limit Exceeded': 'status-mle',
    'Output Limit Exceeded': 'status-ole', 'System Error': 'status-system'
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
.page-header { margin-bottom: 20px; }
.page-desc { color: var(--text-secondary); font-size: 14px; margin-top: 4px; }

.filter-bar {
  display: flex; align-items: center; gap: 10px;
  margin-bottom: 16px; flex-wrap: wrap;
}

.link {
  color: var(--accent-primary); font-weight: 600;
}
.link:hover { text-decoration: underline; }

.sub-id-link {
  color: var(--accent-blue);
  font-family: var(--font-mono);
  font-weight: 600;
  cursor: pointer;
  border-bottom: 1px dashed var(--accent-blue);
  transition: all 0.15s;
}
.sub-id-link:hover {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.pagination-wrap {
  display: flex; justify-content: center; padding: 20px;
}

/* Code Dialog */
.code-dialog-content {
  display: flex; flex-direction: column; gap: 14px;
}
.code-dialog-meta {
  display: flex; align-items: center; gap: 8px;
  font-size: 13px; color: var(--text-secondary);
  flex-wrap: wrap;
}
.meta-sep { color: var(--text-muted); }

.code-dialog-error {
  background: var(--accent-red-bg);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
}
.error-label {
  font-size: 12px; font-weight: 700; color: var(--accent-red);
  margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px;
}
.error-pre {
  margin: 0; white-space: pre-wrap; word-break: break-word;
  font-size: 12px; font-family: var(--font-mono); line-height: 1.5;
}

.code-dialog-code {
  background: var(--code-bg);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  padding: 12px;
}
.code-label {
  font-size: 12px; font-weight: 700; color: var(--text-muted);
  margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px;
}
.code-loading {
  text-align: center; padding: 20px; color: var(--text-muted);
}
.code-block {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 50vh;
  overflow-y: auto;
}
</style>
