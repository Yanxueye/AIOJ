<template>
  <div class="problem-list-page page-container">
    <div class="page-header">
      <h2>题目列表</h2>
    </div>

    <div class="filter-bar card">
      <div class="filter-controls">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索题号或题目名称..."
          prefix-icon="Search"
          clearable
          style="width: 280px"
          @input="debouncedSearch"
        />
        <el-select v-model="filters.difficulty" placeholder="难度" clearable style="width: 120px" @change="loadProblems">
          <el-option label="简单" value="简单" />
          <el-option label="中等" value="中等" />
          <el-option label="困难" value="困难" />
        </el-select>
        <el-select v-model="filters.tag" placeholder="算法标签" clearable style="width: 140px" @change="loadProblems">
          <el-option v-for="t in tags" :key="t" :label="t" :value="t" />
        </el-select>
        <el-select v-model="filters.status" placeholder="做题状态" clearable style="width: 140px" @change="loadProblems">
          <el-option label="已通过" value="accepted" />
          <el-option label="已尝试" value="attempted" />
          <el-option label="未尝试" value="unattempted" />
          <el-option label="已收藏" value="favorite" />
        </el-select>
      </div>
      <el-button v-if="userStore.canManageProblems" type="primary" @click="router.push('/admin/problems/new')">
        添加题目
      </el-button>
    </div>

    <div class="card">
      <el-table
        :data="problemStore.problems"
        v-loading="problemStore.loading"
        stripe
        style="width: 100%"
        @row-click="goToDetail"
        row-class-name="clickable-row"
      >
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.accepted" color="#67c23a" :size="18"><CircleCheckFilled /></el-icon>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="题号" width="80" sortable />
        <el-table-column prop="title" label="题目名称" min-width="200">
          <template #default="{ row }">
            <span class="problem-title-cell">{{ row.title }}</span>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="100" sortable :sort-method="sortByDifficulty">
          <template #default="{ row }">
            <el-tag :type="diffTagType(row.difficulty)" size="small" effect="plain">
              {{ row.difficulty }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分数" prop="difficultyScore" width="90" sortable />
        <el-table-column label="标签" min-width="180">
          <template #default="{ row }">
            <el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag-item" effect="plain" type="info">
              {{ tag }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通过率" width="100" sortable>
          <template #default="{ row }">
            {{ row.acceptRate }}%
          </template>
        </el-table-column>
        <el-table-column v-if="userStore.canManageProblems" label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click.stop="goToEdit(row.id)">修改</el-button>
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

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
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
.page-header h2 {
  font-size: 24px;
  font-weight: 700;
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
  gap: 12px;
  flex-wrap: wrap;
}
.problem-title-cell {
  font-weight: 500;
  color: var(--text-primary);
}
:deep(.clickable-row) {
  cursor: pointer;
}
:deep(.clickable-row:hover .problem-title-cell) {
  color: var(--accent-blue);
}
.tag-item {
  margin-right: 6px;
}
.pagination-wrap {
  display: flex;
  justify-content: center;
  padding-top: 20px;
}
</style>
