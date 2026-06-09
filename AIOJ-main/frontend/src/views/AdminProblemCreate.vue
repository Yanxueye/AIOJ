<template>
  <div class="page-container admin-page">
    <div class="page-header">
      <h2>添加题目</h2>
      <p>在这里填写题面、限制和测试用例，保存后会立即出现在题库列表中。</p>
    </div>

    <ProblemForm ref="problemFormRef" :submitting="submitting" submit-text="创建题目" @submit="handleSubmit" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { problemApi } from '@/api/problem'
import ProblemForm from '@/components/admin/ProblemForm.vue'

const router = useRouter()
const submitting = ref(false)
const problemFormRef = ref(null)

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

async function handleSubmit() {
  const form = problemFormRef.value?.form
  if (!form || !validateForm(form)) {
    return
  }

  submitting.value = true
  try {
    const res = await problemApi.create({
      id: form.id,
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
    ElMessage.success('题目创建成功')
    router.push(`/admin/problems/${res.data.id}/edit`)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.admin-page {
  max-width: 980px;
}
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
</style>
