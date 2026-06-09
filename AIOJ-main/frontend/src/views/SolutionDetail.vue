<template>
  <div class="page-container solution-page" v-loading="loading">
    <div class="page-header">
      <el-button text @click="goBack">返回题解列表</el-button>
      <h2>{{ solution?.title }}</h2>
      <p>{{ solution?.username }} · {{ solution?.language }} · {{ solution?.updatedAt }}</p>
    </div>

    <div class="card">
      <MarkdownRenderer :content="solution?.content || ''" />
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { problemApi } from '@/api/problem'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const solution = ref(null)

onMounted(async () => {
  try {
    const res = await problemApi.getSolutionDetail(route.params.id)
    solution.value = res.data
  } finally {
    loading.value = false
  }
})

function goBack() {
  if (solution.value?.problemId) {
    router.push(`/problem/${solution.value.problemId}`)
    return
  }
  router.push('/my/solutions')
}
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}
.page-header h2 {
  font-size: 28px;
  margin: 10px 0 6px;
}
.page-header p {
  color: var(--text-secondary);
}
</style>
