<template>
  <div class="page-container study-page">
    <div class="page-header">
      <h2>学习计划</h2>
      <p>按主题循序刷题，建立结构化训练路径。</p>
    </div>

    <div class="plan-grid">
      <div v-for="plan in plans" :key="plan.id" class="card plan-card" @click="$router.push(`/study-plans/${plan.id}`)">
        <div class="plan-head">
          <div>
            <div class="plan-title">{{ plan.title }}</div>
            <div class="plan-desc">{{ plan.description }}</div>
          </div>
          <el-tag :type="diffTagType(plan.difficulty)" size="small">{{ plan.difficulty }}</el-tag>
        </div>
        <div class="plan-tags">
          <el-tag v-for="tag in plan.tags" :key="tag" size="small" effect="plain" type="info">{{ tag }}</el-tag>
        </div>
        <div class="plan-meta">
          <span>{{ plan.problemCount }} 题</span>
          <span>已完成 {{ plan.completedCount }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { studyPlanApi } from '@/api/study_plan'

const plans = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await studyPlanApi.getList()
    plans.value = res.data.items || []
  } catch (e) {
    plans.value = []
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
  margin-bottom: 6px;
}
.page-header p {
  color: var(--text-secondary);
}
.plan-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}
.plan-card {
  cursor: pointer;
}
.plan-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}
.plan-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 8px;
}
.plan-desc {
  font-size: 14px;
  color: var(--text-secondary);
}
.plan-tags {
  margin-top: 14px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.plan-meta {
  margin-top: 16px;
  display: flex;
  gap: 16px;
  color: var(--text-muted);
  font-size: 13px;
}
@media (max-width: 900px) {
  .plan-grid {
    grid-template-columns: 1fr;
  }
}
</style>
