<template>
  <div class="page-container detail-page">
    <div v-if="loading" v-loading="true" style="min-height:300px"></div>
    <template v-else-if="plan">
      <div class="page-header">
        <div>
          <h2>{{ plan.title }}</h2>
          <div class="plan-meta">
            <el-tag v-if="plan.difficulty" size="small">{{ plan.difficulty }}</el-tag>
            <span>{{ plan.completedCount || 0 }}/{{ plan.items?.length || 0 }} 完成</span>
          </div>
        </div>
        <el-button @click="$router.back()">返回</el-button>
      </div>
      <p v-if="plan.description" class="plan-desc">{{ plan.description }}</p>
      <div class="problem-list">
        <div v-for="(item,i) in plan.items" :key="item.id" class="problem-row">
          <span class="idx">{{ i+1 }}</span>
          <router-link :to="`/problem/${item.problemId}`" class="title">{{ item.title || `#${item.problemId}` }}</router-link>
          <el-tag v-if="item.difficulty" size="small">{{ item.difficulty }}</el-tag>
        </div>
      </div>
    </template>
    <el-empty v-else description="题单不存在" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import http from '@/api/index'
const route = useRoute()
const plan = ref(null)
const loading = ref(true)
onMounted(async () => {
  try { const r = await http.get(`/study-plans/${route.params.id}`); plan.value = r.data } catch { plan.value = null }
  finally { loading.value = false }
})
</script>

<style scoped>
.detail-page { max-width: 800px; margin: 0 auto }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px }
.page-header h2 { font-size: 22px; font-weight: 800; margin: 0 }
.plan-meta { display: flex; align-items: center; gap: 12px; margin-top: 6px; font-size: 13px; color: var(--text-muted) }
.plan-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.7; margin-bottom: 20px }
.problem-list { display: flex; flex-direction: column; gap: 8px }
.problem-row { display: flex; align-items: center; gap: 12px; padding: 10px 16px; border: 1px solid var(--border-light); border-radius: 8px }
.idx { font-size: 14px; font-weight: 700; color: var(--text-muted); min-width: 24px }
.title { flex: 1; font-weight: 500; text-decoration: none; color: var(--text-primary) }
.title:hover { color: var(--accent-primary) }
</style>
