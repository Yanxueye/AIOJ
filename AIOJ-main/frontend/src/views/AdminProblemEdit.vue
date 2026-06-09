<template>
  <div class="page-container admin-page">
    <div class="page-header">
      <div>
        <h2>编辑题目</h2>
        <p>支持修改题目内容、测试用例和限制，也可以直接删除当前题目。</p>
      </div>
      <el-button @click="router.push('/problems')">返回题库</el-button>
    </div>

    <ProblemForm
      v-loading="problemStore.loading"
      ref="problemFormRef"
      :initial-value="initialValue"
      :submitting="submitting"
      :disable-id="true"
      submit-text="保存修改"
      @submit="handleSubmit"
    >
      <template #actions>
        <el-button type="danger" plain :loading="deleting" @click="handleDelete">删除题目</el-button>
      </template>
    </ProblemForm>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProblemStore } from '@/stores/problem'
import { problemApi } from '@/api/problem'
import ProblemForm from '@/components/admin/ProblemForm.vue'

const route = useRoute()
const router = useRouter()
const problemStore = useProblemStore()
const problemFormRef = ref(null)
const submitting = ref(false)
const deleting = ref(false)

const problemID = computed(() => route.params.id)
const initialValue = computed(() => problemStore.currentProblem || {})

function validateForm(form) {
  if (!form.title.trim() || !form.content.trim()) {
    ElMessage.warning('请填写题目标题和题面')
    return false
  }
  if (form.testCases.some(item => !item.input.trim() || !item.expected.trim())) {
    ElMessage.warning('请完整填写所有测试用例')
    return false
  }
  return true
}

async function loadProblem() {
  await problemStore.fetchAdminProblem(problemID.value)
}

async function handleSubmit() {
  const form = problemFormRef.value?.form
  if (!form || !validateForm(form)) {
    return
  }

  submitting.value = true
  try {
    await problemApi.update(problemID.value, {
      title: form.title,
      difficulty: form.difficulty,
      difficultyScore: form.difficultyScore,
      tags: form.tags,
      source: form.source,
      timeLimit: form.timeLimit,
      memoryLimit: form.memoryLimit,
      outputLimitKb: form.outputLimitKb,
      content: form.content,
      testCases: form.testCases
    })
    ElMessage.success('题目更新成功')
    await loadProblem()
  } finally {
    submitting.value = false
  }
}

async function handleDelete() {
  await ElMessageBox.confirm('删除后无法恢复，确认删除当前题目吗？', '删除题目', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消'
  })

  deleting.value = true
  try {
    await problemApi.remove(problemID.value)
    ElMessage.success('题目已删除')
    router.push('/problems')
  } finally {
    deleting.value = false
  }
}

onMounted(loadProblem)
</script>

<style scoped>
.admin-page {
  max-width: 980px;
}
.page-header {
  margin-bottom: 20px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.page-header h2 {
  font-size: 28px;
  margin-bottom: 6px;
}
.page-header p {
  color: var(--text-secondary);
}
</style>
