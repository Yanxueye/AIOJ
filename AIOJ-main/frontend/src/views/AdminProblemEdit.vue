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
        <el-button type="warning" plain :loading="submitting" @click="handlePublish">发布题目</el-button>
        <el-button type="danger" plain :loading="deleting" @click="handleDelete">删除题目</el-button>
      </template>
    </ProblemForm>

    <div class="card version-card">
      <div class="version-head">重判任务</div>
      <div class="job-actions">
        <el-input v-model="rejudgeReason" placeholder="重判原因，例如更新了隐藏测试或修正了题面" />
        <el-button type="primary" @click="handleRejudge">创建重判任务</el-button>
      </div>
      <el-table :data="rejudgeJobs" size="small" stripe>
        <el-table-column prop="id" label="任务 ID" width="120" />
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column prop="totalSubmissions" label="提交数" width="100" />
        <el-table-column prop="processedCount" label="已处理" width="100" />
        <el-table-column prop="reason" label="原因" min-width="220" />
      </el-table>
    </div>
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
const rejudgeReason = ref('')
const rejudgeJobs = ref([])

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
  const jobs = await problemApi.getRejudgeJobs(problemID.value)
  rejudgeJobs.value = jobs.data.items || []
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
      status: form.status,
      reviewComment: form.reviewComment,
      timeLimit: form.timeLimit,
      memoryLimit: form.memoryLimit,
      outputLimitKb: form.outputLimitKb,
      content: form.content,
      constraints: form.constraints,
      editorial: form.editorial,
      samples: form.samples,
      testCases: form.testCases,
      templates: form.templates
    })
    ElMessage.success('题目更新成功')
    await loadProblem()
  } finally {
    submitting.value = false
  }
}

async function handlePublish() {
  submitting.value = true
  try {
    await problemApi.publish(problemID.value, {
      reviewComment: problemFormRef.value?.form?.reviewComment || ''
    })
    ElMessage.success('题目已发布')
    await loadProblem()
  } finally {
    submitting.value = false
  }
}

async function handleRejudge() {
  await problemApi.rejudge(problemID.value, { reason: rejudgeReason.value })
  ElMessage.success('重判任务已创建')
  rejudgeReason.value = ''
  await loadProblem()
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
.version-card {
  margin-top: 20px;
}
.version-head {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
}
.job-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}
</style>
