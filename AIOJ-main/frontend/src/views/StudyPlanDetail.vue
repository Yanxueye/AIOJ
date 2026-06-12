<template>
  <div class="page-container study-page" v-loading="loading">
    <div class="page-header">
      <el-button text @click="$router.push('/study-plans')">返回学习计划</el-button>
      <h2>{{ plan?.title }}</h2>
      <p>{{ plan?.description }}</p>
      <div class="plan-tags" v-if="plan">
        <el-tag :type="diffTagType(plan.difficulty)" size="small">{{ plan.difficulty }}</el-tag>
        <el-tag v-for="tag in plan.tags" :key="tag" size="small" effect="plain" type="info">{{ tag }}</el-tag>
      </div>
    </div>

    <div class="card">
      <div class="section-title">题单进度</div>
      <div class="progress-line">
        <span>已完成 {{ plan?.completedCount || 0 }} / {{ plan?.items?.length || 0 }}</span>
        <span v-if="plan?.lastCompletedAt" class="progress-time">最近完成：{{ plan.lastCompletedAt }}</span>
        <el-progress :percentage="progressPercent" :stroke-width="10" />
      </div>
    </div>

    <div class="card">
      <div class="section-title">题目列表</div>
      <el-table :data="plan?.items || []" stripe>
        <el-table-column prop="orderNo" label="顺序" width="80" />
        <el-table-column label="题目" min-width="240">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.problemId}`" class="plan-link">
              #{{ row.problemId }} {{ row.title }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column prop="difficulty" label="难度" width="120" />
        <el-table-column label="完成状态" width="140">
          <template #default="{ row }">
            <el-tag :type="row.completed ? 'success' : 'info'" size="small">
              {{ row.completed ? '已完成' : '未完成' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="completedAt" label="完成时间" width="180" />
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { studyPlanApi } from '@/api/study_plan'

const route = useRoute()
const plan = ref(null)
const loading = ref(true)

const progressPercent = computed(() => {
  const total = plan.value?.items?.length || 0
  if (!total) return 0
  return Math.round(((plan.value?.completedCount || 0) / total) * 100)
})

onMounted(async () => {
  try {
    const res = await studyPlanApi.getDetail(route.params.id)
    plan.value = res.data
  } catch (e) {
    plan.value = null
  } finally {
    loading.value = false
  }
})

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
}
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}
.page-header h2 {
  font-size: 28px;
  margin: 8px 0 6px;
}
.page-header p {
  color: var(--text-secondary);
}
.plan-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}
.section-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 12px;
}
.progress-line {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.progress-time {
  font-size: 12px;
  color: var(--text-secondary);
}
.plan-link {
  color: var(--text-primary);
  font-weight: 600;
}
</style>
