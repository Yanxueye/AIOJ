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
        <el-tag v-if="problem" :type="diffTagType(problem.difficulty)" size="small">
          {{ problem.difficulty }}
        </el-tag>
      </div>
      <div class="toolbar-right">
        <el-button :type="problem?.favorite ? 'warning' : 'default'" @click="toggleFavorite">
          <el-icon><Star /></el-icon>{{ problem?.favorite ? '取消收藏' : '收藏题目' }}
        </el-button>
        <el-button :loading="running" @click="handleRun">
          <el-icon><VideoPlay /></el-icon>运行代码
        </el-button>
        <el-button type="success" :loading="submitting" @click="handleSubmit">
          <el-icon><Position /></el-icon>提交代码
        </el-button>
      </div>
    </div>

    <div v-loading="problemStore.loading" class="detail-panels">
      <div class="panel panel-left" :style="{ flex: panelFlex.problem }">
        <div class="panel-tabs">
          <button :class="['tab-btn', { active: activeTab === 'description' }]" @click="activeTab = 'description'">题目描述</button>
          <button :class="['tab-btn', { active: activeTab === 'solutions' }]" @click="activeTab = 'solutions'">题解</button>
          <button :class="['tab-btn', { active: activeTab === 'submissions' }]" @click="activeTab = 'submissions'">提交记录</button>
        </div>

        <div class="panel-body">
          <template v-if="activeTab === 'description'">
            <div class="panel-meta" v-if="problem">
              <el-tag size="small" type="info">时间限制: {{ problem.timeLimit }}ms</el-tag>
              <el-tag size="small" type="info">内存限制: {{ problem.memoryLimit }}MB</el-tag>
              <el-tag size="small" type="info">输出限制: {{ problem.outputLimitKb || 1024 }}KB</el-tag>
            </div>

            <section class="problem-section">
              <MarkdownRenderer v-if="problem?.content" :content="problem.content" />
              <el-empty v-else description="题面内容待补全" :image-size="80" />
            </section>

            <section v-if="problem?.constraints" class="problem-section">
              <div class="section-title">约束条件</div>
              <MarkdownRenderer :content="problem.constraints" />
            </section>

            <section v-if="problem?.samples?.length" class="problem-section">
              <div class="section-title">公开样例</div>
              <div class="sample-list">
                <div v-for="item in problem.samples" :key="item.caseNo" class="sample-card">
                  <div class="sample-head">Sample {{ item.caseNo }}</div>
                  <div class="sample-grid">
                    <div>
                      <div class="sample-label">输入</div>
                      <pre>{{ item.input }}</pre>
                    </div>
                    <div>
                      <div class="sample-label">输出</div>
                      <pre>{{ item.expected }}</pre>
                    </div>
                  </div>
                  <div v-if="item.explanation" class="sample-explain">
                    <div class="sample-label">说明</div>
                    <MarkdownRenderer :content="item.explanation" />
                  </div>
                </div>
              </div>
            </section>

            <section v-if="problem?.relatedProblems?.length" class="problem-section">
              <div class="section-title">相似题推荐</div>
              <div class="related-list">
                <router-link
                  v-for="item in problem.relatedProblems"
                  :key="item.id"
                  :to="`/problem/${item.id}`"
                  class="related-item"
                >
                  <div class="related-title">#{{ item.id }} {{ item.title }}</div>
                  <div class="related-meta">
                    <el-tag :type="diffTagType(item.difficulty)" size="small" effect="plain">{{ item.difficulty }}</el-tag>
                    <el-tag v-for="tag in item.tags" :key="tag" size="small" type="info" effect="plain">{{ tag }}</el-tag>
                  </div>
                </router-link>
              </div>
            </section>
          </template>

          <template v-else-if="activeTab === 'solutions'">
            <section class="problem-section">
              <div class="section-title">官方题解</div>
              <MarkdownRenderer v-if="problem?.editorial" :content="problem.editorial" />
              <el-empty v-else description="暂无官方题解" :image-size="80" />
            </section>

            <section class="problem-section">
              <div class="section-title section-row">
                <span>我的题解</span>
                <el-button type="primary" plain size="small" @click="goToMySolutionEditor">
                  {{ problem?.mySolution?.id ? '编辑我的题解' : '新增题解' }}
                </el-button>
              </div>
              <el-empty description="点击按钮进入独立题解编辑页" :image-size="70" />
            </section>

            <section class="problem-section">
              <div class="section-title">用户题解</div>
              <div v-if="problem?.solutions?.length" class="solution-list">
                <router-link v-for="item in problem.solutions" :key="item.id" :to="`/solutions/${item.id}`" class="solution-item solution-link-card">
                  <div class="solution-head">
                    <div>
                      <div class="solution-title">{{ item.title }}</div>
                      <div class="solution-meta">{{ item.username }} · {{ item.language }} · {{ item.updatedAt }}</div>
                    </div>
                  </div>
                  <div class="solution-preview">{{ item.content.slice(0, 160) }}<span v-if="item.content.length > 160">...</span></div>
                </router-link>
              </div>
              <el-empty v-else description="还没有已发布题解" :image-size="80" />
            </section>
          </template>

          <template v-else>
            <section class="problem-section">
              <div class="section-title">最近提交</div>
              <div v-if="recentSubmissions.length" class="submission-list">
                <div v-for="item in recentSubmissions" :key="item.id" class="submission-item">
                  <div class="submission-main">
                    <span class="submission-id">#{{ item.id }}</span>
                    <span :class="statusClass(item.status)">{{ item.status }}</span>
                  </div>
                  <div class="submission-meta">
                    <span>{{ item.language }}</span>
                    <span>{{ item.runtimeMs }}ms</span>
                    <span>{{ item.memoryKb }} KB</span>
                  </div>
                </div>
              </div>
              <el-empty v-else description="还没有提交记录" :image-size="80" />
            </section>
          </template>
        </div>
      </div>

      <div class="divider" @mousedown="e => startResize(e, 'left')" />

      <div class="panel panel-editor" :style="{ flex: panelFlex.editor }">
        <CodeEditor
          ref="codeEditorRef"
          v-model="code"
          :language="language"
          :templates="templateMap"
          :draft-key="draftKey"
          :legacy-draft-key="legacyDraftKey"
          @change-language="lang => language = lang"
        />

        <div class="run-panel">
          <div class="run-panel-head">
            <span>自定义输入</span>
            <span class="run-panel-tip">用于 Run Code，不会计入正式提交记录</span>
          </div>
          <el-input
            v-model="customInput"
            type="textarea"
            :rows="4"
            placeholder="输入自定义测试数据"
            resize="none"
          />
        </div>

        <transition name="slide-up">
          <div v-if="resultView" class="result-panel">
            <div class="result-header">
              <div class="result-title">
                <span class="result-source">{{ resultView.source === 'run' ? '运行结果' : '提交结果' }}</span>
                <span :class="statusClass(resultView.status)">
                  {{ resultView.status }}
                </span>
              </div>
              <el-button text size="small" @click="resultView = null">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>

            <div class="result-details">
              <span v-if="resultView.traceId">Trace: {{ resultView.traceId }}</span>
              <span>运行时间: {{ displayRuntime(resultView) }}</span>
              <span>内存: {{ displayMemory(resultView) }}</span>
              <span v-if="resultView.finishedAt">完成时间: {{ formatTime(resultView.finishedAt) }}</span>
            </div>

            <div v-if="resultView.errorMessage" class="result-block">
              <div class="result-block-title">错误信息</div>
              <pre>{{ resultView.errorMessage }}</pre>
            </div>

            <div v-if="resultView.compileOutput" class="result-block">
              <div class="result-block-title">编译输出</div>
              <pre>{{ resultView.compileOutput }}</pre>
            </div>

            <div v-if="resultView.stdout || resultView.stderr" class="result-block">
              <div class="result-block-title">运行输出</div>
              <pre v-if="resultView.stdout">{{ resultView.stdout }}</pre>
              <pre v-if="resultView.stderr" class="case-error">{{ resultView.stderr }}</pre>
            </div>

            <div v-if="resultView.caseResults?.length" class="result-block">
              <div class="result-block-title">测试点结果</div>
              <div class="case-list">
                <div v-for="item in resultView.caseResults" :key="item.caseNo" class="case-item">
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
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useProblemStore } from '@/stores/problem'
import { useSubmissionStore } from '@/stores/submission'
import { useUserStore } from '@/stores/user'
import CodeEditor from '@/components/CodeEditor.vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const route = useRoute()
const router = useRouter()
const problemStore = useProblemStore()
const submissionStore = useSubmissionStore()
const userStore = useUserStore()

const problem = computed(() => problemStore.currentProblem)
const recentSubmissions = ref([])
const activeTab = ref('description')

const templateMap = computed(() => {
  const entries = Array.isArray(problem.value?.templates) ? problem.value.templates : []
  return entries.reduce((acc, item) => {
    if (item.language && item.code) {
      acc[item.language] = item.code
    }
    return acc
  }, {})
})

const code = ref('')
const language = ref('cpp')
const customInput = ref('')
const submitting = ref(false)
const running = ref(false)
const resultView = ref(null)
const codeEditorRef = ref(null)

const draftNamespace = computed(() => userStore.userInfo?.id ? `user-${userStore.userInfo.id}` : 'guest')
const draftKey = computed(() => `${draftNamespace.value}:problem-${route.params.id}`)
const legacyDraftKey = computed(() => `problem-${route.params.id}`)

const panelFlex = reactive({
  problem: 1.15,
  editor: 1,
})

let resizeState = null

function startResize(e) {
  e.preventDefault()
  resizeState = { startX: e.clientX, startFlex: { ...panelFlex } }
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

function onResize(e) {
  if (!resizeState) return
  const dx = e.clientX - resizeState.startX
  const scale = dx / window.innerWidth * 3
  panelFlex.problem = Math.max(0.4, resizeState.startFlex.problem + scale)
  panelFlex.editor = Math.max(0.4, resizeState.startFlex.editor - scale)
}

function stopResize() {
  resizeState = null
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
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

async function hydrateSubmissionResult(result) {
  if (!result?.id || result.source !== 'submit') {
    resultView.value = result
    return result
  }
  const [casesRes, outputRes] = await Promise.all([
    submissionStore.getCases(result.id),
    submissionStore.getOutput(result.id)
  ])
  const hydrated = {
    ...result,
    caseResults: casesRes?.items || result.caseResults || [],
    stdout: outputRes?.stdout || '',
    stderr: outputRes?.stderr || ''
  }
  resultView.value = hydrated
  return hydrated
}

async function handleSubmit() {
  const codeVal = code.value || codeEditorRef.value?.getCode()
  if (!codeVal?.trim()) {
    ElMessage.warning('请先输入代码')
    return
  }
  submitting.value = true
  resultView.value = null
  try {
    const result = await submissionStore.submit({
      problemId: problem.value.id,
      language: language.value,
      code: codeVal
    })
    const hydrated = await hydrateSubmissionResult(result)
    if (hydrated.status === 'Accepted') {
      ElMessage.success('通过')
    } else {
      ElMessage.warning(`评测结果: ${hydrated.status}`)
    }
    await loadRecentSubmissions()
  } catch {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

async function handleRun() {
  const codeVal = code.value || codeEditorRef.value?.getCode()
  if (!codeVal?.trim()) {
    ElMessage.warning('请先输入代码')
    return
  }
  running.value = true
  resultView.value = null
  try {
    const res = await problemStore.runProblem(problem.value.id, {
      language: language.value,
      code: codeVal,
      customInput: customInput.value
    })
    resultView.value = res
    if (res.status === 'Accepted') {
      ElMessage.success('运行完成')
    } else {
      ElMessage.warning(`运行结果: ${res.status}`)
    }
  } catch {
    ElMessage.error('运行失败')
  } finally {
    running.value = false
  }
}

async function toggleFavorite() {
  if (!problem.value?.id) return
  if (problem.value.favorite) {
    await problemStore.unfavoriteProblem(problem.value.id)
    problem.value.favorite = false
    ElMessage.success('已取消收藏')
  } else {
    await problemStore.favoriteProblem(problem.value.id)
    problem.value.favorite = true
    ElMessage.success('已收藏题目')
  }
}

async function loadRecentSubmissions() {
  const res = await submissionStore.fetchSubmissions({
    page: 1,
    pageSize: 5,
    problemId: problem.value?.id || ''
  })
  recentSubmissions.value = submissionStore.submissions.slice(0, 5)
  return res
}

function goToMySolutionEditor() {
  const my = problem.value?.mySolution
  if (my?.id) {
    router.push(`/my/solutions/${my.id}/edit`)
    return
  }
  ElMessage.info('请先在“我的题解”页创建一篇题解草稿。')
}

watch(() => route.params.id, async () => {
  language.value = 'cpp'
  code.value = ''
  customInput.value = ''
  resultView.value = null
  activeTab.value = 'description'
  const loaded = await problemStore.fetchProblem(route.params.id)
  await loadRecentSubmissions()
})

onMounted(async () => {
  const loaded = await problemStore.fetchProblem(route.params.id)
  await loadRecentSubmissions()
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
  min-width: 220px;
}
.panel-left {
  background: #fff;
}
.panel-tabs {
  display: flex;
  gap: 4px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color);
  background: #fafbfc;
}
.tab-btn {
  border: none;
  background: transparent;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
}
.tab-btn.active {
  background: #edf4ff;
  color: var(--accent-blue);
  font-weight: 600;
}
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
.panel-meta {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.problem-section + .problem-section {
  margin-top: 28px;
}
.section-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 12px;
}
.section-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.sample-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.sample-card {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 14px;
  background: #fafbfc;
}
.sample-head {
  font-weight: 700;
  margin-bottom: 10px;
}
.sample-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.sample-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}
.sample-card pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 12px;
}
.sample-explain {
  margin-top: 10px;
}
.solution-editor,
.solution-list,
.submission-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.solution-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.solution-tip {
  font-size: 12px;
  color: var(--text-secondary);
}
.solution-item,
.submission-item {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 14px;
  background: #fafbfc;
}
.solution-link-card {
  display: block;
  text-decoration: none;
  color: inherit;
}
.solution-link-card:hover {
  border-color: var(--accent-blue);
  background: #f0f7ff;
}
.solution-head {
  margin-bottom: 10px;
}
.solution-title {
  font-size: 16px;
  font-weight: 700;
}
.solution-meta,
.submission-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.solution-preview {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
}
.submission-main {
  display: flex;
  gap: 12px;
  align-items: center;
}
.submission-id {
  font-family: monospace;
  color: var(--text-muted);
}
.related-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.related-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  background: #fafbfc;
  text-decoration: none;
}
.related-item:hover {
  border-color: var(--accent-blue);
  background: #f0f7ff;
}
.related-title {
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-primary);
}
.related-meta {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.panel-editor {
  background: #1e1e2e;
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
.run-panel {
  background: #161621;
  border-top: 1px solid #2d2d3f;
  padding: 12px 14px;
}
.run-panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #e5e7eb;
  font-size: 13px;
  margin-bottom: 8px;
}
.run-panel-tip {
  color: #9ca3af;
  font-size: 12px;
}
.run-panel :deep(.el-textarea__inner) {
  background: #0f172a;
  color: #e5e7eb;
  border-color: #374151;
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
.result-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.result-source {
  color: var(--text-secondary);
  font-size: 13px;
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
@media (max-width: 1200px) {
  .sample-grid {
    grid-template-columns: 1fr;
  }
}
</style>
