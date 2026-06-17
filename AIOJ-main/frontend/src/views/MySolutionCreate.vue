<template>
  <div class="page-container solution-page">
    <div class="page-header">
      <el-button text @click="$router.back()">返回</el-button>
      <h2>新建题解</h2>
      <p>先选择题目，再编写题解内容。</p>
    </div>

    <div class="card form-card">
      <el-input-number v-model="form.problemId" :min="1" placeholder="题号" />
      <el-input v-model="form.title" placeholder="题解标题" />
      <el-input v-model="form.language" placeholder="题解语言，如 cpp / python / go" />
      <div class="editor-toolbar">
        <el-button type="primary" plain size="small" :loading="aiLoading" @click="generateAISolution">
          <el-icon><MagicStick /></el-icon> AI 辅助生成
        </el-button>
        <span class="ai-tip">基于你的提交历史自动生成题解草稿</span>
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
import { reactive, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { problemApi } from '@/api/problem'
import { aiApi } from '@/api/ai'

const route = useRoute()
const router = useRouter()
const saving = ref(false)
const aiLoading = ref(false)
const form = reactive({
  problemId: 0,
  title: '',
  content: '',
  language: 'cpp',
  isPublished: false
})

onMounted(() => {
  const pid = route.query.problemId
  if (pid) {
    form.problemId = Number(pid)
  }
})

async function generateAISolution() {
  if (!form.problemId) {
    ElMessage.warning('请先填写题号')
    return
  }
  aiLoading.value = true
  try {
    // Fetch problem info for title
    const problemRes = await problemApi.getDetail(form.problemId)
    const problem = problemRes.data

    // Get user's latest AC code for this problem
    let code = '', language = form.language
    try {
      const { submissionApi } = await import('@/api/submission')
      const subRes = await submissionApi.getList({ problemId: form.problemId, status: 'Accepted', pageSize: 1 })
      const ac = subRes.data?.list?.[0]
      if (ac) {
        const detail = await submissionApi.getDetail(ac.id)
        code = detail.data?.code || ''
        language = detail.data?.language || language
      }
    } catch { /* ignore */ }

    const res = await aiApi.generateSolution({
      problemId: form.problemId,
      language,
      code
    })
    const data = res.data
    const content = data?.content || data?.rawMarkdown || ''
    if (content) {
      form.content = content
      if (!form.title && data?.title) {
        form.title = data.title
      } else if (!form.title) {
        form.title = `${problem.title || '题解'} - AI 辅助生成`
      }
      ElMessage.success('AI 题解已生成，请检查并修改后保存')
    } else {
      ElMessage.warning('AI 未能生成题解，请手动编写')
    }
  } catch {
    ElMessage.error('AI 服务暂时不可用')
  } finally {
    aiLoading.value = false
  }
}

async function handleSave() {
  if (!form.problemId) {
    ElMessage.warning('请先填写题号')
    return
  }
  saving.value = true
  try {
    await problemApi.saveSolution(form.problemId, {
      title: form.title,
      content: form.content,
      language: form.language,
      isPublished: form.isPublished
    })
    ElMessage.success(form.isPublished ? '题解已发布' : '题解草稿已保存')
    router.push(`/problem/${form.problemId}`)
  } finally {
    saving.value = false
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
