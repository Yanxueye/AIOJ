<template>
  <div class="page-container solution-page" v-loading="loading">
    <div class="page-header">
      <el-button text @click="goBack">返回</el-button>
      <h2>{{ form.title || '编辑题解' }}</h2>
      <p>{{ problemTitle }}</p>
    </div>

    <div class="card form-card">
      <el-input v-model="form.title" placeholder="题解标题" />
      <el-input v-model="form.language" placeholder="题解语言，如 cpp / python / go" />
      <div class="editor-toolbar">
        <el-button type="primary" plain size="small" :loading="aiLoading" @click="regenerateAI">
          <el-icon><MagicStick /></el-icon> AI 辅助重写
        </el-button>
        <span class="ai-tip">基于你的提交历史重新生成题解</span>
      </div>
      <el-input v-model="form.content" type="textarea" :rows="18" placeholder="输入你的题解内容（支持 Markdown）" />
      <div class="form-actions">
        <el-switch v-model="form.isPublished" active-text="发布题解" inactive-text="仅保存草稿" />
        <el-button type="primary" :loading="saving" @click="handleSave">保存题解</el-button>
      </div>
      <div class="tip">只有通过该题后才能发布题解，未通过时只能保存草稿。</div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { problemApi } from '@/api/problem'
import { aiApi } from '@/api/ai'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const aiLoading = ref(false)
const problemTitle = ref('')
const form = reactive({
  id: 0,
  problemId: 0,
  title: '',
  content: '',
  language: '',
  isPublished: false
})

onMounted(async () => {
  try {
    const res = await problemApi.getMySolutionDetail(route.params.id)
    const data = res.data
    form.id = data.id
    form.problemId = data.problemId
    form.title = data.title || ''
    form.content = data.content || ''
    form.language = data.language || ''
    form.isPublished = Boolean(data.isPublished)
    problemTitle.value = data.problemTitle ? `题目：#${data.problemId} ${data.problemTitle}` : `题目：#${data.problemId}`
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  if (!form.problemId) return
  saving.value = true
  try {
    await problemApi.saveSolution(form.problemId, {
      title: form.title,
      content: form.content,
      language: form.language,
      isPublished: form.isPublished
    })
    ElMessage.success(form.isPublished ? '题解已发布' : '题解草稿已保存')
  } finally {
    saving.value = false
  }
}

function goBack() {
  if (form.problemId) {
    router.push(`/problem/${form.problemId}`)
  } else {
    router.back()
  }
}

async function regenerateAI() {
  if (!form.problemId) return
  aiLoading.value = true
  try {
    const res = await aiApi.solveProblem({
      problemId: form.problemId,
      question: '请帮我生成一篇题解，包含：1) 解题思路概述 2) 踩坑点 3) 实现亮点 4) 关键公式/算法 5) 时间/空间复杂度。用Markdown格式输出。',
      level: 'full'
    })
    const answer = res.data?.answer || ''
    if (answer) {
      form.content = answer
      ElMessage.success('AI 题解已生成，请检查后保存')
    } else {
      ElMessage.warning('AI 未能生成题解')
    }
  } catch {
    ElMessage.error('AI 服务暂时不可用')
  } finally {
    aiLoading.value = false
  }
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
.form-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}
.ai-tip {
  font-size: 12px;
  color: var(--text-muted);
}
.form-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}
.tip {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
