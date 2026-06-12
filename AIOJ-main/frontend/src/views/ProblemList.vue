<template>
  <div class="problem-list-page page-container">
    <div class="page-header">
      <h2>题目列表</h2>
      <p class="page-desc">精选算法题目，支持按难度、标签和状态筛选</p>
    </div>

    <div class="filter-bar card">
      <div class="filter-controls">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索题号或题目名称..."
          prefix-icon="Search"
          clearable
          class="search-input"
          @input="debouncedSearch"
        />
        <el-select v-model="filters.difficulty" placeholder="难度" clearable style="width: 110px" @change="loadProblems">
          <el-option label="简单" value="简单" />
          <el-option label="中等" value="中等" />
          <el-option label="困难" value="困难" />
        </el-select>
        <el-select v-model="filters.tag" placeholder="算法标签" clearable style="width: 130px" @change="loadProblems">
          <el-option v-for="t in tags" :key="t" :label="t" :value="t" />
        </el-select>
        <el-select v-model="filters.status" placeholder="做题状态" clearable style="width: 130px" @change="loadProblems">
          <el-option label="已通过" value="accepted" />
          <el-option label="已尝试" value="attempted" />
          <el-option label="未尝试" value="unattempted" />
          <el-option label="已收藏" value="favorite" />
        </el-select>
      </div>
      <el-button v-if="userStore.isAdmin" type="primary" round @click="router.push('/admin/problems/new')">
        <el-icon><Plus /></el-icon>添加题目
      </el-button>
    </div>

    <div class="card table-card">
      <el-table
        :data="problemStore.problems"
        v-loading="problemStore.loading"
        style="width: 100%"
        @row-click="goToDetail"
        row-class-name="clickable-row"
        :header-cell-style="{ background: 'var(--bg-warm)', color: 'var(--text-secondary)', fontWeight: '600', fontSize: '12.5px', textTransform: 'uppercase', letterSpacing: '0.04em' }"
      >
        <el-table-column label="状态" width="60" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.accepted" :style="{ color: 'var(--accent-green)' }" :size="18"><CircleCheckFilled /></el-icon>
            <span v-else class="status-dash">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="题号" width="80" sortable>
          <template #default="{ row }">
            <span class="problem-id-cell">#{{ row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="题目名称" min-width="200">
          <template #default="{ row }">
            <span class="problem-title-cell">{{ row.title }}</span>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="90" sortable :sort-method="sortByDifficulty">
          <template #default="{ row }">
            <span :class="diffClass(row.difficulty)" class="difficulty-label">{{ row.difficulty }}</span>
          </template>
        </el-table-column>
        <el-table-column label="分数" prop="difficultyScore" width="80" sortable>
          <template #default="{ row }">
            <span class="score-cell">{{ row.difficultyScore }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="180">
          <template #default="{ row }">
            <el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag-item" effect="plain" type="info">
              {{ tag }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通过率" width="90" sortable>
          <template #default="{ row }">
            <span class="rate-cell">{{ row.acceptRate }}%</span>
          </template>
        </el-table-column>
        <el-table-column v-if="userStore.isAdmin" label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click.stop="goToEdit(row.id)">修改</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="problemStore.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadProblems"
          @size-change="loadProblems"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useProblemStore } from '@/stores/problem'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const problemStore = useProblemStore()
const userStore = useUserStore()

const tags = ['动态规划', '贪心', '搜索', '图论', '数学', '字符串', '数据结构', '模拟', '排序', '二分']

const filters = reactive({ keyword: '', difficulty: '', tag: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 20 })

let searchTimer = null
function debouncedSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadProblems(), 300)
}

function loadProblems() {
  problemStore.fetchProblems({
    page: pagination.page,
    pageSize: pagination.pageSize,
    keyword: filters.keyword,
    difficulty: filters.difficulty,
    tag: filters.tag,
    status: filters.status
  })
}

function goToDetail(row) {
  router.push(`/problem/${row.id}`)
}

function goToEdit(id) {
  router.push(`/admin/problems/${id}/edit`)
}

function diffClass(d) {
  return d === '简单' ? 'diff-easy' : d === '中等' ? 'diff-medium' : 'diff-hard'
}

function sortByDifficulty(a, b) {
  const order = { '简单': 1, '中等': 2, '困难': 3 }
  return (order[a.difficulty] || 0) - (order[b.difficulty] || 0)
}

onMounted(async () => {
  if (userStore.isLoggedIn) {
    try {
      await userStore.fetchProfile()
    } catch {}
  }
  loadProblems()
})
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.search-input {
  width: 260px;
}

.table-card {
  padding: 0;
  overflow: hidden;
}

.table-card :deep(.el-table) {
  --el-table-border-color: var(--border-light);
  --el-table-tr-bg-color: transparent;
  --el-table-row-hover-bg-color: var(--bg-active);
}

.table-card :deep(.el-table td) {
  border-bottom: 1px solid var(--border-light);
}

.problem-id-cell {
  font-family: var(--font-mono);
  font-size: 12.5px;
  color: var(--text-muted);
  font-weight: 500;
}

.problem-title-cell {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 14px;
}

.status-dash {
  color: var(--text-muted);
  font-size: 14px;
}

.diff-easy {
  color: var(--accent-green);
  font-weight: 700;
  font-size: 13px;
}
.diff-medium {
  color: var(--accent-orange);
  font-weight: 700;
  font-size: 13px;
}
.diff-hard {
  color: var(--accent-red);
  font-weight: 700;
  font-size: 13px;
}

.score-cell {
  font-family: var(--font-mono);
  font-size: 12.5px;
  color: var(--text-secondary);
}

.rate-cell {
  font-size: 13px;
  color: var(--text-secondary);
}

:deep(.clickable-row) {
  cursor: pointer;
}
:deep(.clickable-row:hover .problem-title-cell) {
  color: var(--accent-primary-dark);
}

.tag-item {
  margin-right: 6px;
  margin-bottom: 2px;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px;
}

@media (max-width: 768px) {
  .search-input {
    width: 100%;
  }
}
</style>
