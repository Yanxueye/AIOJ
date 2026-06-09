<template>
  <div class="problem-detail-page">
    <div class="detail-toolbar">
      <div class="toolbar-left">
        <el-button text @click="$router.push('/problems')">
          <el-icon><ArrowLeft /></el-icon>返回题目列表
        </el-button>
        <el-divider direction="vertical" />
        <span v-if="problem" class="problem-id">#{{ problem.id }}</span>
        <span v-if="problem" class="problem-name">{{ problem.title }}</span>
        <el-tag v-if="problem" :type="diffTagType(problem.difficulty)" size="small" style="margin-left: 8px">
          {{ problem.difficulty }}
        </el-tag>
      </div>
      <div class="toolbar-right">
        <el-tooltip content="切换 AI 助手面板">
          <el-button :type="showAI ? 'primary' : 'default'" circle size="small" @click="showAI = !showAI">
            <el-icon><MagicStick /></el-icon>
          </el-button>
        </el-tooltip>
        <el-button type="success" :loading="submitting" @click="handleSubmit">
          <el-icon><Position /></el-icon>提交代码
        </el-button>
      </div>
    </div>

    <div v-loading="problemStore.loading" class="detail-panels">
      <div class="panel panel-problem" :style="{ flex: panelFlex.problem }">
        <div class="panel-header">
          <span>题目描述</span>
          <div class="panel-meta" v-if="problem">
            <el-tag size="small" type="info">时间限制: {{ problem.timeLimit }}ms</el-tag>
            <el-tag size="small" type="info">内存限制: {{ problem.memoryLimit }}MB</el-tag>
            <el-tag size="small" type="info">输出限制: {{ problem.outputLimitKb || 1024 }}KB</el-tag>
          </div>
        </div>
        <div class="panel-content">
          <MarkdownRenderer v-if="problem" :content="problem.content" />
        </div>
      </div>

      <div class="divider" @mousedown="e => startResize(e, 'left')" />

      <div class="panel panel-editor" :style="{ flex: panelFlex.editor }">
        <CodeEditor
          ref="codeEditorRef"
          v-model="code"
          :language="language"
          :draft-key="`problem-${route.params.id}`"
          @change-language="lang => language = lang"
        />
        <transition name="slide-up">
          <div v-if="submissionResult" class="result-panel">
            <div class="result-header">
              <span :class="statusClass(submissionResult.status)">
                {{ submissionResult.status }}
              </span>
              <el-button text size="small" @click="submissionResult = null">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>

            <div class="result-details">
              <span v-if="submissionResult.traceId">Trace: {{ submissionResult.traceId }}</span>
              <span>运行时间: {{ displayRuntime(submissionResult) }}</span>
              <span>内存: {{ displayMemory(submissionResult) }}</span>
              <span v-if="submissionResult.finishedAt">完成时间: {{ formatTime(submissionResult.finishedAt) }}</span>
            </div>

            <div v-if="submissionResult.errorMessage" class="result-block">
              <div class="result-block-title">错误信息</div>
              <pre>{{ submissionResult.errorMessage }}</pre>
            </div>

            <div v-if="submissionResult.compileOutput" class="result-block">
              <div class="result-block-title">编译输出</div>
              <pre>{{ submissionResult.compileOutput }}</pre>
            </div>

            <div v-if="submissionResult.caseResults?.length" class="result-block">
              <div class="result-block-title">测试点结果</div>
              <div class="case-list">
                <div v-for="item in submissionResult.caseResults" :key="item.caseNo" class="case-item">
                  <div class="case-top">
                    <span>Case {{ item.caseNo }}</span>
                    <span :class="statusClass(item.status)">{{ item.status }}</span>
                  </div>
                  <div class="case-meta">
                    <span>{{ item.runtimeMs ?? 0 }} ms</span>
                    <span>{{ item.memoryKb ?? 0 }} KB</span>
                    <span>{{ item.stdoutBytes ?? 0 }} stdoutB</span>
                    <span>{{ item.stderrBytes ?? 0 }} stderrB</span>
                    <span v-if="item.signal">signal: {{ item.signal }}</span>
                  </div>
                  <pre v-if="item.stdoutPreview" class="case-preview">{{ item.stdoutPreview }}</pre>
                  <pre v-if="item.stderrPreview" class="case-preview case-error">{{ item.stderrPreview }}</pre>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <template v-if="showAI">
        <div class="divider" @mousedown="e => startResize(e, 'right')" />
        <div class="panel panel-ai" :style="{ flex: panelFlex.ai }">
          <AIChat
            :problem-context="problem"
            :code-context="{ code, language }"
            @clear-context="() => {}"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useProblemStore } from '@/stores/problem'
import { useSubmissionStore } from '@/stores/submission'
import { useAIStore } from '@/stores/ai'
import { ElMessage } from 'element-plus'
import CodeEditor from '@/components/CodeEditor.vue'
import AIChat from '@/components/AIChat.vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const route = useRoute()
const problemStore = useProblemStore()
const submissionStore = useSubmissionStore()
const aiStore = useAIStore()

const problem = computed(() => problemStore.currentProblem)
const showAI = ref(false)
const code = ref('')
const language = ref('cpp')
const submitting = ref(false)
const submissionResult = ref(null)
const codeEditorRef = ref(null)

const panelFlex = reactive({
  problem: 1,
  editor: 1,
  ai: 0.8
})

let resizeState = null

function startResize(e, side) {
  e.preventDefault()
  resizeState = { side, startX: e.clientX, startFlex: { ...panelFlex } }
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

function onResize(e) {
  if (!resizeState) return
  const dx = e.clientX - resizeState.startX
  const scale = dx / window.innerWidth * 3

  if (resizeState.side === 'left') {
    panelFlex.problem = Math.max(0.3, resizeState.startFlex.problem + scale)
    panelFlex.editor = Math.max(0.3, resizeState.startFlex.editor - scale)
  } else {
    panelFlex.editor = Math.max(0.3, resizeState.startFlex.editor + scale)
    panelFlex.ai = Math.max(0.3, resizeState.startFlex.ai - scale)
  }
}

function stopResize() {
  resizeState = null
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
}

async function handleSubmit() {
  const codeVal = code.value || codeEditorRef.value?.getCode()
  if (!codeVal?.trim()) {
    ElMessage.warning('请先输入代码')
    return
  }
  submitting.value = true
  submissionResult.value = null
  try {
    const result = await submissionStore.submit({
      problemId: problem.value.id,
      language: language.value,
      code: codeVal
    })
    submissionResult.value = result
    if (result.status === 'Accepted') {
      ElMessage.success('通过')
    } else {
      ElMessage.warning(`评测结果: ${result.status}`)
    }
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

function statusClass(status) {
  const map = {
    Pending: 'status-pending',
    Queueing: 'status-pending',
    Compiling: 'status-running',
    Running: 'status-running',
    Accepted: 'status-accepted',
    'Wrong Answer': 'status-wrong',
    'Compile Error': 'status-ce',
    'Runtime Error': 'status-wrong',
    'Time Limit Exceeded': 'status-tle',
    'Memory Limit Exceeded': 'status-mle',
    'Output Limit Exceeded': 'status-ole',
    'System Error': 'status-system'
  }
  return map[status] || ''
}

function displayRuntime(result) {
  const runtime = result?.runtimeMs ?? result?.runtime
  return runtime != null ? `${runtime}ms` : '-'
}

function displayMemory(result) {
  if (result?.memoryKb != null && result.memoryKb > 0) {
    return `${result.memoryKb} KB`
  }
  if (result?.memory != null) {
    return `${result.memory} MB`
  }
  return '-'
}

function formatTime(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
}

onMounted(async () => {
  const loaded = await problemStore.fetchProblem(route.params.id)
  aiStore.startNewConversation(loaded)
})

onBeforeUnmount(() => {
  stopResize()
})
</script>

<style scoped>
.problem-detail-page {
  height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
}
.detail-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.problem-id {
  font-family: monospace;
  color: var(--text-muted);
  font-size: 14px;
}
.problem-name {
  font-size: 16px;
  font-weight: 600;
}

.detail-panels {
  flex: 1;
  display: flex;
  overflow: hidden;
  background: var(--bg-primary);
}
.panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 200px;
}
.panel-problem {
  background: #fff;
}
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid var(--border-color);
  background: #fafbfc;
  flex-shrink: 0;
}
.panel-meta {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
.panel-editor {
  background: #1e1e2e;
}
.panel-ai {
  min-width: 280px;
}

.divider {
  width: 5px;
  cursor: col-resize;
  background: var(--border-color);
  transition: background 0.2s;
  flex-shrink: 0;
}
.divider:hover {
  background: var(--accent-blue);
}

.result-panel {
  background: #fff;
  border-top: 2px solid var(--border-color);
  padding: 12px 16px;
  flex-shrink: 0;
  max-height: 46vh;
  overflow-y: auto;
}
.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  font-weight: 700;
}
.result-details {
  display: flex;
  gap: 24px;
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  flex-wrap: wrap;
}
.result-block {
  margin-top: 12px;
}
.result-block-title {
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 6px;
}
.result-block pre,
.case-preview {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f5f7fa;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 12px;
  line-height: 1.5;
}
.case-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.case-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  background: #fafbfc;
}
.case-top,
.case-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.case-top {
  font-weight: 600;
}
.case-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}
.case-preview {
  margin-top: 8px;
}
.case-error {
  background: #fff2f0;
}

.slide-up-enter-active, .slide-up-leave-active {
  transition: all 0.3s ease;
}
.slide-up-enter-from, .slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
