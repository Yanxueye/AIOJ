<template>
  <div class="problem-detail-page">
    <!-- Toolbar -->
    <div class="detail-toolbar">
      <div class="toolbar-left">
        <el-button text @click="$router.push('/problems')">
          <el-icon><ArrowLeft /></el-icon>返回
        </el-button>
        <el-divider direction="vertical" />
        <span v-if="problem" class="problem-id">#{{ problem.id }}</span>
        <span v-if="problem" class="problem-title">{{ problem.title }}</span>
        <el-tag v-if="problem" :type="diffTagType(problem.difficulty)" size="small" effect="plain">
          {{ problem.difficulty }}
        </el-tag>
      </div>
      <div class="toolbar-right">
        <el-button :type="problem?.favorite ? 'warning' : 'default'" text @click="toggleFavorite">
          <el-icon><Star /></el-icon>
        </el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          提交
        </el-button>
      </div>
    </div>

    <!-- Main panels -->
    <div v-loading="problemStore.loading" class="detail-panels">
      <!-- LEFT PANEL -->
      <div class="panel panel-left" :style="{ flex: panelFlex.left }">
        <!-- Tab bar: show normal tabs OR result tab -->
        <div class="panel-tabs">
          <template v-if="!resultView">
            <button :class="['tab-btn', { active: leftTab === 'description' }]" @click="leftTab = 'description'">题目描述</button>
            <button :class="['tab-btn', { active: leftTab === 'solutions' }]" @click="leftTab = 'solutions'">题解</button>
            <button :class="['tab-btn', { active: leftTab === 'submissions' }]" @click="leftTab = 'submissions'">提交记录</button>
          </template>
          <template v-else>
            <button class="tab-btn active result-tab">
              <span :class="statusIconClass(resultView.status)" class="result-status-icon">●</span>
              {{ resultView.source === 'run' ? '运行结果' : '提交结果' }}
            </button>
            <button class="tab-btn close-result" @click="resultView = null">
              <el-icon><Close /></el-icon>
            </button>
          </template>
        </div>

        <!-- Left panel body -->
        <div class="panel-body">
          <!-- DESCRIPTION TAB -->
          <template v-if="!resultView && leftTab === 'description'">
            <div class="panel-meta" v-if="problem">
              <el-tag size="small" type="info">时间: {{ problem.timeLimit }}ms</el-tag>
              <el-tag size="small" type="info">内存: {{ problem.memoryLimit }}MB</el-tag>
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
              <div class="section-title section-row" @click="samplesCollapsed = !samplesCollapsed" style="cursor: pointer; user-select: none">
                <span>示例</span>
                <el-icon class="collapse-arrow" :class="{ rotated: samplesCollapsed }"><ArrowDown /></el-icon>
              </div>
              <div v-show="!samplesCollapsed" class="sample-list">
                <div v-for="item in problem.samples" :key="item.caseNo" class="sample-card">
                  <div class="sample-head">示例 {{ item.caseNo }}</div>
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
              <div class="section-title">相似题目</div>
              <div class="related-list">
                <router-link v-for="item in problem.relatedProblems" :key="item.id" :to="`/problem/${item.id}`" class="related-item">
                  <span class="related-title">#{{ item.id }} {{ item.title }}</span>
                  <el-tag :type="diffTagType(item.difficulty)" size="small" effect="plain">{{ item.difficulty }}</el-tag>
                </router-link>
              </div>
            </section>
          </template>

          <!-- SOLUTIONS TAB -->
          <template v-else-if="!resultView && leftTab === 'solutions'">
            <section class="problem-section">
              <div class="section-title section-row">
                <span>题解 ({{ allSolutions.length }})</span>
                <el-button type="primary" plain size="small" @click="goToMySolutionEditor">
                  <el-icon><EditPen /></el-icon>发布题解
                </el-button>
              </div>
              <div v-if="allSolutions.length" class="solution-list">
                <div v-for="item in allSolutions" :key="item.id" class="solution-item" @click="viewSolution(item)">
                  <div class="solution-head">
                    <div class="solution-title-row">
                      <el-tag v-if="item.isOfficial" type="success" size="small" effect="dark">官方</el-tag>
                      <span class="solution-title">{{ item.title }}</span>
                    </div>
                    <div class="solution-meta">
                      <span>{{ item.username }}</span>
                      <span>{{ item.language }}</span>
                      <span>{{ item.updatedAt }}</span>
                    </div>
                  </div>
                  <div class="solution-preview">{{ item.content.slice(0, 150) }}<span v-if="item.content.length > 150">...</span></div>
                  <div class="solution-actions">
                    <el-button v-if="item.id > 0" size="small" :type="item.liked ? 'primary' : 'default'" text @click.stop="handleLike(item)">
                      <el-icon><Star /></el-icon> {{ item.likeCount || 0 }}
                    </el-button>
                    <el-button v-if="item.id > 0 && item.userId === userStore.userInfo?.id" size="small" text @click.stop="goToEditSolution(item.id)">
                      <el-icon><Edit /></el-icon>编辑
                    </el-button>
                    <el-button v-if="item.id > 0 && userStore.isAdmin && item.userId !== userStore.userInfo?.id" size="small" text type="danger" @click.stop="handleDeleteSolution(item)">
                      <el-icon><Delete /></el-icon>删除
                    </el-button>
                  </div>
                </div>
              </div>
              <el-empty v-else description="还没有题解" :image-size="80" />
            </section>
            <!-- Solution dialog -->
            <el-dialog v-model="solutionDialogVisible" :title="viewingSolution?.title" width="700px" top="5vh">
              <div v-if="viewingSolution" class="solution-dialog-content">
                <div class="solution-dialog-meta">
                  <el-tag v-if="viewingSolution.isOfficial" type="success" size="small" effect="dark">官方题解</el-tag>
                  <span>{{ viewingSolution.username }} · {{ viewingSolution.language }} · {{ viewingSolution.updatedAt }}</span>
                  <div style="margin-left: auto">
                    <el-button size="small" :type="viewingSolution.liked ? 'primary' : 'default'" text @click="handleLike(viewingSolution)">
                      <el-icon><Star /></el-icon> {{ viewingSolution.likeCount || 0 }}
                    </el-button>
                  </div>
                </div>
                <MarkdownRenderer :content="viewingSolution.content" />
              </div>
            </el-dialog>
          </template>

          <!-- SUBMISSIONS TAB -->
          <template v-else-if="!resultView && leftTab === 'submissions'">
            <section class="problem-section">
              <div class="section-title section-row">
                <span>提交记录</span>
                <router-link to="/status" class="view-all-link">查看全部 →</router-link>
              </div>
              <div v-if="recentSubmissions.length" class="submission-list">
                <div v-for="item in recentSubmissions" :key="item.id" class="submission-item" @click="toggleSubmissionCode(item)">
                  <div class="submission-main">
                    <span :class="statusIconClass(item.status)" class="result-status-icon">●</span>
                    <span :class="statusTextClass(item.status)">{{ item.status }}</span>
                    <span class="submission-lang">{{ item.language }}</span>
                    <span class="submission-time">{{ item.runtimeMs ?? '-' }}ms</span>
                    <span class="submission-mem">{{ item.memoryKb ? (item.memoryKb > 1024 ? (item.memoryKb / 1024).toFixed(1) + 'MB' : item.memoryKb + 'KB') : '-' }}</span>
                    <el-icon class="expand-icon"><ArrowDown v-if="!expandedSubmissions[item.id]" /><ArrowUp v-else /></el-icon>
                  </div>
                  <transition name="fade">
                    <div v-if="expandedSubmissions[item.id]" class="submission-code-block" @click.stop>
                      <div v-if="loadingCode[item.id]" class="code-loading"><el-icon class="is-loading"><Loading /></el-icon> 加载中...</div>
                      <pre v-else class="submission-code"><code>{{ submissionCodes[item.id] || '暂无代码' }}</code></pre>
                    </div>
                  </transition>
                </div>
              </div>
              <el-empty v-else description="还没有提交记录" :image-size="80" />
            </section>
          </template>

          <!-- RESULT VIEW (replaces left panel when active) -->
          <template v-else-if="resultView">
            <div class="result-view">
              <!-- Status banner -->
              <div class="result-status-banner" :class="getBannerClass(resultView.status)">
                <div class="banner-status">
                  <span class="banner-status-text">{{ resultView.status }}</span>
                  <span v-if="resultView.status === 'Accepted'" class="banner-check">✓</span>
                  <span v-else-if="['Pending', 'Queueing', 'Compiling', 'Running'].includes(resultView.status)" class="banner-spinner" />
                </div>
                <div class="banner-stats">
                  <div class="banner-stat">
                    <span class="banner-stat-label">执行用时</span>
                    <span class="banner-stat-value">{{ displayRuntime(resultView) }}</span>
                  </div>
                  <div class="banner-stat">
                    <span class="banner-stat-label">内存消耗</span>
                    <span class="banner-stat-value">{{ displayMemory(resultView) }}</span>
                  </div>
                  <div class="banner-stat">
                    <span class="banner-stat-label">语言</span>
                    <span class="banner-stat-value">{{ resultView.language }}</span>
                  </div>
                </div>
              </div>

              <!-- Error details -->
              <div v-if="resultView.status !== 'Accepted'" class="result-error-section">
                <div v-if="resultView.errorMessage" class="error-block">
                  <div class="error-block-title">错误信息</div>
                  <pre class="error-pre">{{ resultView.errorMessage }}</pre>
                </div>
                <div v-if="resultView.compileOutput" class="error-block">
                  <div class="error-block-title">编译输出</div>
                  <pre class="error-pre">{{ resultView.compileOutput }}</pre>
                </div>
                <div v-if="resultView.stdout || resultView.stderr" class="error-block">
                  <div class="error-block-title">运行输出</div>
                  <pre v-if="resultView.stdout" class="error-pre">{{ resultView.stdout }}</pre>
                  <pre v-if="resultView.stderr" class="error-pre error-text">{{ resultView.stderr }}</pre>
                </div>
              </div>

              <!-- Failed case details -->
              <div v-if="failedCases.length" class="result-cases-section">
                <div class="cases-title">未通过的测试用例</div>
                <div v-for="item in failedCases" :key="item.caseNo" class="case-card">
                  <div class="case-header">
                    <span>测试用例 {{ item.caseNo }}</span>
                    <span :class="statusTextClass(item.status)">{{ item.status }}</span>
                  </div>
                  <div class="case-body">
                    <div class="case-field">
                      <div class="case-field-label">输入</div>
                      <pre class="case-field-value">{{ getCaseInput(item) }}</pre>
                    </div>
                    <div class="case-compare">
                      <div class="case-field">
                        <div class="case-field-label expected-label">预期输出</div>
                        <pre class="case-field-value expected-value">{{ getCaseExpected(item) }}</pre>
                      </div>
                      <div class="case-field">
                        <div class="case-field-label actual-label">实际输出</div>
                        <pre class="case-field-value actual-value">{{ item.stdoutPreview || '(无输出)' }}</pre>
                      </div>
                    </div>
                    <div v-if="item.stderrPreview" class="case-field">
                      <div class="case-field-label">错误信息</div>
                      <pre class="case-field-value error-text">{{ item.stderrPreview }}</pre>
                    </div>
                  </div>
                  <div class="case-meta">
                    <span>{{ item.runtimeMs ?? 0 }}ms</span>
                    <span>{{ item.memoryKb ?? 0 }}KB</span>
                  </div>
                </div>
              </div>

              <!-- AI Analysis -->
              <div class="result-ai-section">
                <div class="ai-section-header">
                  <span class="ai-section-title">
                    <el-icon><MagicStick /></el-icon>
                    AI 分析
                  </span>
                  <el-button
                    v-if="!aiAnalysis && !aiAnalysisLoading"
                    type="primary"
                    size="small"
                    @click="fetchAIAnalysis"
                  >
                    {{ resultView.status === 'Accepted' ? '分析代码' : '获取诊断' }}
                  </el-button>
                </div>
                <div v-if="aiAnalysisLoading" class="ai-loading">
                  <el-icon class="is-loading"><Loading /></el-icon>
                  <span>AI 正在分析代码...</span>
                </div>
                <div v-else-if="aiAnalysis" class="ai-content">
                  <MarkdownRenderer :content="aiAnalysis" />
                </div>
                <div v-else class="ai-placeholder">
                  点击按钮，AI 将分析你的代码并给出{{ resultView.status === 'Accepted' ? '优化建议' : '改进建议' }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- RESIZE DIVIDER -->
      <div class="divider" @mousedown="e => startResize(e)" />

      <!-- RIGHT PANEL: Editor + Test Cases -->
      <div class="panel panel-right" :style="{ flex: panelFlex.right }">
        <!-- Code Editor -->
        <div class="editor-area">
          <CodeEditor
            ref="codeEditorRef"
            v-model="code"
            :language="language"
            :templates="templateMap"
            :draft-key="draftKey"
            :legacy-draft-key="legacyDraftKey"
            @change-language="lang => language = lang"
          />
        </div>

        <!-- Test Case Panel -->
        <div class="testcase-panel" :style="{ height: testcasePanelHeight + 'px' }">
          <!-- Resizable divider -->
          <div class="panel-resize-handle" @mousedown.prevent="startTestcaseResize">
            <div class="resize-dots" />
          </div>

          <!-- Header -->
          <div class="testcase-header">
            <div class="testcase-header-left">
              <div class="testcase-tabs">
                <button
                  v-for="(tc, idx) in testCases"
                  :key="idx"
                  :class="['tc-tab', { active: activeCaseIdx === idx }]"
                  @click="activeCaseIdx = idx"
                >{{ tc.label }}</button>
                <button class="tc-tab tc-add" @click="addTestCase" title="添加测试用例">+</button>
              </div>
            </div>
            <div class="testcase-header-right">
              <el-button type="primary" size="small" :loading="running" @click="runCurrentCase" class="run-btn">
                <el-icon><VideoPlay /></el-icon>运行
              </el-button>
              <div class="collapse-btn" @click="testcaseCollapsed = true" title="收起">
                <el-icon><ArrowDown /></el-icon>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div v-if="!testcaseCollapsed" class="testcase-body">
            <div class="tc-section">
              <div class="tc-label">输入</div>
              <textarea
                v-model="testCases[activeCaseIdx].input"
                class="tc-textarea"
                rows="3"
                spellcheck="false"
                placeholder="输入测试数据..."
              />
            </div>
            <div v-if="testCases[activeCaseIdx].expected !== undefined" class="tc-section">
              <div class="tc-label">
                预期输出
                <button class="tc-toggle" @click="toggleExpected">隐藏</button>
              </div>
              <textarea
                v-model="testCases[activeCaseIdx].expected"
                class="tc-textarea"
                rows="2"
                spellcheck="false"
                placeholder="预期输出（可选）"
              />
            </div>
            <div v-else>
              <button class="tc-toggle" @click="toggleExpected">+ 添加预期输出</button>
            </div>

            <!-- Run result inline -->
            <div v-if="resultView && resultView.source === 'run'" class="tc-run-result">
              <div class="tc-run-header">
                <span class="tc-run-status" :class="statusBadgeClass(resultView.status)">
                  {{ resultView.status }}
                </span>
                <span class="tc-run-stats">{{ displayRuntime(resultView) }} | {{ displayMemory(resultView) }}</span>
              </div>
              <div v-if="resultView.stdout" class="tc-run-output">
                <div class="tc-label">输出</div>
                <pre class="tc-pre">{{ resultView.stdout }}</pre>
              </div>
              <div v-if="resultView.stderr" class="tc-run-output">
                <div class="tc-label">错误</div>
                <pre class="tc-pre error-text">{{ resultView.stderr }}</pre>
              </div>
              <div v-if="resultView.compileOutput" class="tc-run-output">
                <div class="tc-label">编译输出</div>
                <pre class="tc-pre">{{ resultView.compileOutput }}</pre>
              </div>
            </div>
          </div>

          <!-- Collapsed state -->
          <div v-else class="testcase-collapsed" @click="testcaseCollapsed = false">
            <span class="collapsed-hint">点击展开测试用例</span>
            <el-icon class="expand-icon"><ArrowUp /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- AI Chat FAB -->
    <div class="ai-chat-fab" @click="chatOpen = !chatOpen">
      <el-icon :size="22"><MagicStick /></el-icon>
    </div>
    <transition name="slide-right">
      <div v-if="chatOpen" class="ai-chat-panel">
        <div class="ai-chat-header">
          <span><el-icon><MagicStick /></el-icon> AI 助手</span>
          <el-button text size="small" @click="chatOpen = false"><el-icon><Close /></el-icon></el-button>
        </div>
        <div class="ai-chat-messages" ref="chatMessagesRef">
          <div v-for="(msg, i) in chatMessages" :key="i" :class="['chat-msg', msg.role]">
            <div class="chat-msg-content">{{ msg.content }}</div>
          </div>
          <div v-if="chatLoading" class="chat-msg assistant">
            <div class="chat-msg-content"><el-icon class="is-loading"><Loading /></el-icon> 思考中...</div>
          </div>
        </div>
        <div class="ai-chat-input">
          <el-input v-model="chatInput" placeholder="询问关于这道题的问题..." @keyup.enter="sendChat" :disabled="chatLoading">
            <template #append>
              <el-button @click="sendChat" :loading="chatLoading">发送</el-button>
            </template>
          </el-input>
          <div class="chat-context-hint">
            <el-button text size="small" @click="sendCodeContext">
              <el-icon><Document /></el-icon> 发送当前代码
            </el-button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProblemStore } from '@/stores/problem'
import { problemApi } from '@/api/problem'
import { aiApi } from '@/api/ai'
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
const leftTab = ref('description')
const expandedSubmissions = ref({})
const submissionCodes = ref({})
const loadingCode = ref({})
const solutionDialogVisible = ref(false)
const viewingSolution = ref(null)
const aiAnalysis = ref('')
const aiAnalysisLoading = ref(false)
const chatOpen = ref(false)
const chatMessages = ref([{ role: 'assistant', content: '你好！我是 AI 助手，可以帮你解答关于这道题的问题、分析代码或给出提示。' }])
const chatInput = ref('')
const chatLoading = ref(false)
const chatMessagesRef = ref(null)
const testcaseCollapsed = ref(false)
const samplesCollapsed = ref(false)
const testcasePanelHeight = ref(260)
let testcaseResizeState = null

function startTestcaseResize(e) {
  e.preventDefault()
  const panelRight = document.querySelector('.panel-right')
  const maxH = panelRight ? panelRight.clientHeight - 100 : 600
  testcaseResizeState = { startY: e.clientY, startHeight: testcasePanelHeight.value, maxH }
  document.addEventListener('mousemove', onTestcaseResize)
  document.addEventListener('mouseup', stopTestcaseResize)
}
function onTestcaseResize(e) {
  if (!testcaseResizeState) return
  const dy = testcaseResizeState.startY - e.clientY
  const newH = testcaseResizeState.startHeight + dy
  if (newH < 60) {
    testcasePanelHeight.value = 120
    testcaseCollapsed.value = true
    stopTestcaseResize()
    return
  }
  testcasePanelHeight.value = Math.max(120, Math.min(testcaseResizeState.maxH, newH))
}
function stopTestcaseResize() {
  testcaseResizeState = null
  document.removeEventListener('mousemove', onTestcaseResize)
  document.removeEventListener('mouseup', stopTestcaseResize)
}

const bannerClassMap = { Accepted: 'banner-accepted', Pending: 'banner-pending', Queueing: 'banner-pending', Compiling: 'banner-pending', Running: 'banner-pending' }
function getBannerClass(status) { return bannerClassMap[status] || 'banner-error' }

function statusBadgeClass(status) {
  if (status === 'Accepted') return 'badge-accepted'
  if (['Wrong Answer', 'Runtime Error'].includes(status)) return 'badge-error'
  if (status === 'Compile Error') return 'badge-ce'
  if (status === 'Time Limit Exceeded') return 'badge-tle'
  if (status === 'Memory Limit Exceeded') return 'badge-mle'
  if (status === 'Output Limit Exceeded') return 'badge-ole'
  if (status === 'System Error') return 'badge-system'
  if (['Pending', 'Queueing'].includes(status)) return 'badge-pending'
  if (['Compiling', 'Running'].includes(status)) return 'badge-running'
  return 'badge-default'
}

const templateMap = computed(() => {
  const entries = Array.isArray(problem.value?.templates) ? problem.value.templates : []
  return entries.reduce((acc, item) => {
    if (item.language && item.code) acc[item.language] = item.code
    return acc
  }, {})
})

const code = ref('')
const language = ref('cpp')
const submitting = ref(false)
const running = ref(false)
const resultView = ref(null)
const codeEditorRef = ref(null)

// Test case management
const testCases = ref([{ label: '用例 1', input: '', expected: undefined }])
const activeCaseIdx = ref(0)
const customInput = computed(() => testCases.value[activeCaseIdx.value]?.input || '')

function addTestCase() {
  testCases.value.push({ label: `用例 ${testCases.value.length + 1}`, input: '', expected: undefined })
  activeCaseIdx.value = testCases.value.length - 1
}
function toggleExpected() {
  const tc = testCases.value[activeCaseIdx.value]
  tc.expected = tc.expected === undefined ? '' : undefined
}
function loadSamples() {
  if (problem.value?.samples?.length) {
    testCases.value = problem.value.samples.map((s, i) => ({
      label: `样例 ${i + 1}`, input: s.input || '', expected: s.expected || ''
    }))
    activeCaseIdx.value = 0
  }
}

// Failed cases for result view
const failedCases = computed(() => {
  if (!resultView.value?.caseResults?.length) return []
  // Show only the first failed case (LeetCode-style)
  const first = resultView.value.caseResults.find(c => c.status !== 'Accepted')
  return first ? [first] : []
})
function getCaseInput(item) {
  return item.input || problem.value?.samples?.[item.caseNo - 1]?.input || '(隐藏用例)'
}
function getCaseExpected(item) {
  return item.expected || problem.value?.samples?.[item.caseNo - 1]?.expected || '(隐藏用例)'
}

const draftNamespace = computed(() => userStore.userInfo?.id ? `user-${userStore.userInfo.id}` : 'guest')
const draftKey = computed(() => `${draftNamespace.value}:problem-${route.params.id}`)
const legacyDraftKey = computed(() => `problem-${route.params.id}`)

const panelFlex = reactive({ left: 1.15, right: 1 })
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
  panelFlex.left = Math.max(0.4, resizeState.startFlex.left + scale)
  panelFlex.right = Math.max(0.4, resizeState.startFlex.right - scale)
}
function stopResize() {
  resizeState = null
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
}

function statusIconClass(status) {
  if (status === 'Accepted') return 'icon-accepted'
  if (['Wrong Answer', 'Runtime Error'].includes(status)) return 'icon-wrong'
  if (['Compile Error'].includes(status)) return 'icon-ce'
  if (['Time Limit Exceeded'].includes(status)) return 'icon-tle'
  if (['Memory Limit Exceeded'].includes(status)) return 'icon-mle'
  if (['Output Limit Exceeded'].includes(status)) return 'icon-ole'
  if (['System Error'].includes(status)) return 'icon-system'
  return 'icon-pending'
}
function statusTextClass(status) {
  const map = {
    Pending: 'text-pending', Queueing: 'text-pending', Compiling: 'text-running', Running: 'text-running',
    Accepted: 'text-accepted', 'Wrong Answer': 'text-wrong', 'Compile Error': 'text-ce',
    'Runtime Error': 'text-wrong', 'Time Limit Exceeded': 'text-tle',
    'Memory Limit Exceeded': 'text-mle', 'Output Limit Exceeded': 'text-ole', 'System Error': 'text-system'
  }
  return map[status] || ''
}
function diffTagType(d) { return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger' }
function displayRuntime(result) {
  const runtime = result?.runtimeMs ?? result?.runtime
  return runtime != null ? `${runtime}ms` : '-'
}
function displayMemory(result) {
  if (result?.memoryKb != null && result.memoryKb > 0) return `${result.memoryKb} KB`
  if (result?.memory != null) return `${result.memory} MB`
  return '-'
}
function formatTime(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

async function toggleSubmissionCode(item) {
  const id = item.id
  if (expandedSubmissions.value[id]) { expandedSubmissions.value[id] = false; return }
  expandedSubmissions.value[id] = true
  if (submissionCodes.value[id]) return
  loadingCode.value[id] = true
  try {
    const { submissionApi } = await import('@/api/submission')
    const res = await submissionApi.getDetail(id)
    submissionCodes.value[id] = res.data?.code || ''
  } catch { submissionCodes.value[id] = '加载失败' }
  finally { loadingCode.value[id] = false }
}

async function hydrateSubmissionResult(result) {
  if (!result?.id || result.source !== 'submit') { resultView.value = result; return result }
  const [casesRes, outputRes] = await Promise.all([submissionStore.getCases(result.id), submissionStore.getOutput(result.id)])
  const hydrated = { ...result, caseResults: casesRes?.items || result.caseResults || [], stdout: outputRes?.stdout || '', stderr: outputRes?.stderr || '' }
  resultView.value = hydrated
  return hydrated
}

async function handleSubmit() {
  const codeVal = code.value || codeEditorRef.value?.getCode()
  if (!codeVal?.trim()) { ElMessage.warning('请先输入代码'); return }
  submitting.value = true
  aiAnalysis.value = ''
  // Show pending state immediately
  resultView.value = { status: 'Pending', source: 'submit', language: language.value }
  try {
    const result = await submissionStore.submit({ problemId: problem.value.id, language: language.value, code: codeVal })
    // Update result view with each intermediate status during polling
    if (result) {
      const hydrated = await hydrateSubmissionResult(result)
      if (hydrated.status === 'Accepted') ElMessage.success('通过')
      else ElMessage.warning(`评测结果: ${hydrated.status}`)
    }
    await loadRecentSubmissions()
  } catch { ElMessage.error('提交失败') }
  finally { submitting.value = false }
}

async function runCurrentCase() { await handleRun() }
async function handleRun() {
  const codeVal = code.value || codeEditorRef.value?.getCode()
  if (!codeVal?.trim()) { ElMessage.warning('请先输入代码'); return }
  running.value = true
  // For run, show result inline in test case panel (not left panel)
  const prevResult = resultView.value
  if (resultView.value?.source !== 'submit') resultView.value = null
  try {
    const res = await problemStore.runProblem(problem.value.id, { language: language.value, code: codeVal, customInput: customInput.value })
    resultView.value = { ...res, source: 'run' }
    if (res.status === 'Accepted') ElMessage.success('运行完成')
    else ElMessage.warning(`运行结果: ${res.status}`)
  } catch { ElMessage.error('运行失败'); resultView.value = prevResult }
  finally { running.value = false }
}

async function toggleFavorite() {
  if (!problem.value?.id) return
  if (problem.value.favorite) { await problemStore.unfavoriteProblem(problem.value.id); problem.value.favorite = false; ElMessage.success('已取消收藏') }
  else { await problemStore.favoriteProblem(problem.value.id); problem.value.favorite = true; ElMessage.success('已收藏题目') }
}

async function loadRecentSubmissions() {
  const res = await submissionStore.fetchSubmissions({ page: 1, pageSize: 5, problemId: problem.value?.id || '' })
  recentSubmissions.value = submissionStore.submissions.slice(0, 5)
  return res
}

function goToMySolutionEditor() {
  const my = problem.value?.mySolution
  if (my?.id) { router.push(`/my/solutions/${my.id}/edit`); return }
  router.push(`/my/solutions/new?problemId=${route.params.id}`)
}

const likedSet = ref(new Set())
const allSolutions = computed(() => (problem.value?.solutions || []).map(s => ({ ...s, liked: likedSet.value.has(s.id) })))
function viewSolution(item) { viewingSolution.value = item; solutionDialogVisible.value = true }
async function handleLike(item) {
  try {
    const res = await problemApi.likeSolution(item.id)
    const newLiked = res.data?.liked ?? !item.liked
    item.liked = newLiked
    item.likeCount = res.data?.likeCount ?? (newLiked ? item.likeCount + 1 : Math.max(0, item.likeCount - 1))
    if (newLiked) likedSet.value.add(item.id); else likedSet.value.delete(item.id)
  } catch { ElMessage.error('操作失败') }
}
function goToEditSolution(id) { router.push(`/my/solutions/${id}/edit`) }
async function handleDeleteSolution(item) {
  try {
    await ElMessageBox.confirm('确定要删除这篇题解吗？', '删除题解', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
    await problemApi.deleteSolution(item.id)
    ElMessage.success('题解已删除')
    await problemStore.fetchProblem(route.params.id)
  } catch (e) { if (e !== 'cancel') ElMessage.error('删除失败') }
}

function isTerminalStatus(status) {
  return ['Accepted', 'Wrong Answer', 'Compile Error', 'Runtime Error', 'Time Limit Exceeded', 'Memory Limit Exceeded', 'Output Limit Exceeded', 'System Error'].includes(status)
}

async function fetchAIAnalysis() {
  if (!resultView.value || !code.value) return
  aiAnalysisLoading.value = true; aiAnalysis.value = ''
  try {
    const rv = resultView.value
    const payload = {
      problemId: problem.value.id,
      language: language.value,
      code: code.value,
      judgeStatus: rv.status,
      errorMessage: rv.errorMessage || rv.compileOutput || '',
      runtimeMs: rv.runtimeMs ?? rv.runtime ?? 0,
      memoryKb: rv.memoryKb ?? 0
    }
    const res = await aiApi.diagnoseCode(payload)
    aiAnalysis.value = res.data?.rawMarkdown || res.data?.summary || '暂时无法提供分析'
  } catch { aiAnalysis.value = 'AI 分析暂时不可用' }
  finally { aiAnalysisLoading.value = false }
}

function scrollChatToBottom() { setTimeout(() => { if (chatMessagesRef.value) chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight }, 50) }
async function sendChat() {
  const msg = chatInput.value.trim()
  if (!msg || chatLoading.value) return
  chatMessages.value.push({ role: 'user', content: msg }); chatInput.value = ''; chatLoading.value = true; scrollChatToBottom()
  try {
    const res = await aiApi.chat({ message: msg, problem_id: problem.value?.id, history: chatMessages.value.slice(-10).map(m => ({ role: m.role, content: m.content })) })
    chatMessages.value.push({ role: 'assistant', content: res.data?.reply || '暂时无法回复' })
  } catch { chatMessages.value.push({ role: 'assistant', content: 'AI 服务暂时不可用' }) }
  finally { chatLoading.value = false; scrollChatToBottom() }
}
async function sendCodeContext() {
  if (!code.value) { ElMessage.warning('编辑器中没有代码'); return }
  chatInput.value = `请分析我的代码并给出改进建议：\n\`\`\`${language.value}\n${code.value}\n\`\`\``
  sendChat()
}

// Sync submission status updates to resultView during active submission
watch(() => submissionStore.currentResult, (newResult) => {
  if (submitting.value && newResult && resultView.value?.source === 'submit') {
    resultView.value = { ...resultView.value, ...newResult, source: 'submit' }
  }
}, { deep: true })

watch(() => route.params.id, async () => {
  language.value = 'cpp'; code.value = ''; resultView.value = null; leftTab.value = 'description'
  await problemStore.fetchProblem(route.params.id); loadSamples(); await loadRecentSubmissions()
})
onMounted(async () => { await problemStore.fetchProblem(route.params.id); loadSamples(); await loadRecentSubmissions() })
onBeforeUnmount(() => { stopResize() })
</script>

<style scoped>
/* Layout */
.problem-detail-page { height: calc(100vh - 60px); display: flex; flex-direction: column; background: var(--bg-primary); }
.detail-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 0 16px; height: 46px; background: var(--bg-card); border-bottom: 1px solid var(--border-light); flex-shrink: 0; }
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: 8px; }
.problem-id { font-family: 'SF Mono', 'Cascadia Code', monospace; color: var(--text-muted); font-size: 13px; }
.problem-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.detail-panels { flex: 1; display: flex; overflow: hidden; }
.panel { display: flex; flex-direction: column; overflow: hidden; min-width: 200px; }
.panel-left { background: var(--bg-card); }
.panel-right { display: flex; flex-direction: column; background: var(--editor-bg); }

/* Tabs */
.panel-tabs { display: flex; align-items: center; height: 40px; padding: 0 12px; border-bottom: 1px solid var(--border-light); background: var(--bg-card); flex-shrink: 0; gap: 2px; }
.tab-btn { border: none; background: transparent; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); transition: all 0.15s; white-space: nowrap; }
.tab-btn:hover { color: var(--text-primary); background: var(--bg-hover); }
.tab-btn.active { color: var(--text-primary); font-weight: 600; background: var(--bg-hover); }
.tab-btn.result-tab { display: flex; align-items: center; gap: 6px; }
.tab-btn.close-result { margin-left: auto; padding: 4px 8px; color: var(--text-muted); }

/* Status icons */
.result-status-icon { font-size: 10px; }
.icon-accepted { color: var(--accent-green); }
.icon-wrong { color: var(--accent-red); }
.icon-ce { color: var(--accent-purple); }
.icon-tle { color: var(--accent-orange); }
.icon-mle { color: #c2410c; }
.icon-pending { color: var(--text-muted); }
.text-accepted { color: var(--accent-green); font-weight: 600; }
.text-wrong { color: var(--accent-red); font-weight: 600; }
.text-ce { color: var(--accent-purple); font-weight: 600; }
.text-tle { color: var(--accent-orange); font-weight: 600; }
.text-mle { color: #c2410c; font-weight: 600; }
.text-ole { color: #b45309; font-weight: 600; }
.text-system { color: #7c3aed; font-weight: 600; }
.text-running { color: var(--accent-blue); font-weight: 600; }
.text-pending { color: var(--text-muted); }

/* Left panel body */
.panel-body { flex: 1; overflow-y: auto; padding: 16px 20px; }
.problem-section + .problem-section { margin-top: 24px; }
.section-title { font-size: 15px; font-weight: 700; margin-bottom: 10px; }
.section-row { display: flex; align-items: center; justify-content: space-between; }
.collapse-arrow { font-size: 14px; color: var(--text-muted); transition: transform 0.2s; }
.collapse-arrow.rotated { transform: rotate(-90deg); }
.section-row { display: flex; align-items: center; justify-content: space-between; }
.panel-meta { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 14px; }

/* Samples */
.sample-list { display: flex; flex-direction: column; gap: 12px; }
.sample-card { border: 1px solid var(--border-light); border-radius: 8px; padding: 12px; background: var(--bg-hover); }
.sample-head { font-weight: 600; margin-bottom: 8px; font-size: 13px; }
.sample-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.sample-label { font-size: 11px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.5px; }
.sample-card pre { margin: 0; white-space: pre-wrap; word-break: break-word; background: var(--bg-card); border: 1px solid var(--border-light); border-radius: 6px; padding: 8px 10px; font-size: 12px; font-family: 'SF Mono', 'Cascadia Code', monospace; }
.sample-explain { margin-top: 8px; }

/* Solutions */
.solution-list { display: flex; flex-direction: column; gap: 10px; }
.solution-item { border: 1px solid var(--border-light); border-radius: 8px; padding: 12px; background: var(--bg-hover); cursor: pointer; transition: border-color 0.15s; }
.solution-item:hover { border-color: var(--accent-primary); }
.solution-head { margin-bottom: 6px; }
.solution-title { font-size: 14px; font-weight: 600; }
.solution-title-row { display: flex; align-items: center; gap: 6px; }
.solution-meta { margin-top: 2px; font-size: 11px; color: var(--text-muted); display: flex; gap: 10px; }
.solution-preview { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.solution-actions { margin-top: 6px; display: flex; gap: 4px; }
.solution-dialog-content { max-height: 60vh; overflow-y: auto; }
.solution-dialog-meta { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--text-muted); margin-bottom: 14px; }

/* Submissions */
.submission-list { display: flex; flex-direction: column; gap: 6px; }
.submission-item { border: 1px solid var(--border-light); border-radius: 8px; padding: 10px 12px; background: var(--bg-hover); cursor: pointer; transition: border-color 0.15s; }
.submission-item:hover { border-color: var(--accent-primary); }
.submission-main { display: flex; align-items: center; gap: 10px; font-size: 13px; }
.submission-lang, .submission-time, .submission-mem { font-size: 12px; color: var(--text-muted); }
.expand-icon { margin-left: auto; font-size: 12px; color: var(--text-muted); }
.submission-code-block { margin-top: 10px; border-top: 1px solid var(--border-light); padding-top: 10px; }
.submission-code { background: var(--code-bg); border-radius: 6px; padding: 10px; font-size: 12px; line-height: 1.6; overflow-x: auto; max-height: 300px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; margin: 0; }
.code-loading { text-align: center; padding: 12px; color: var(--text-muted); font-size: 12px; }
.related-list { display: flex; flex-direction: column; gap: 6px; }
.related-item { display: flex; align-items: center; justify-content: space-between; padding: 8px 10px; border-radius: 6px; background: var(--bg-hover); text-decoration: none; color: inherit; transition: background 0.15s; }
.related-item:hover { background: var(--accent-primary-bg); }
.related-title { font-size: 13px; font-weight: 500; }
.view-all-link { font-size: 12px; color: var(--accent-blue); }
.view-all-link:hover { text-decoration: underline; }

/* Result View (left panel) */
.result-view { display: flex; flex-direction: column; gap: 16px; }
.result-status-banner { border-radius: 10px; padding: 20px; }
.banner-accepted { background: linear-gradient(135deg, #f0fdf4, #dcfce7); border: 1px solid #86efac; }
.banner-pending { background: linear-gradient(135deg, #fffbeb, #fef3c7); border: 1px solid #fcd34d; }
.banner-error { background: linear-gradient(135deg, #fef2f2, #fee2e2); border: 1px solid #fca5a5; }
[data-theme="dark"] .banner-accepted { background: linear-gradient(135deg, #052e16, #14532d); border-color: #16a34a; }
[data-theme="dark"] .banner-pending { background: linear-gradient(135deg, #451a03, #78350f); border-color: #d97706; }
[data-theme="dark"] .banner-error { background: linear-gradient(135deg, #450a0a, #7f1d1d); border-color: #dc2626; }
.banner-status { display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
.banner-status-text { font-size: 22px; font-weight: 800; letter-spacing: -0.02em; }
.banner-check { font-size: 24px; color: var(--accent-green); }
.banner-spinner {
  width: 18px; height: 18px; flex-shrink: 0;
  border: 2.5px solid #d97706;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.banner-stats { display: flex; gap: 28px; }
.banner-stat { display: flex; flex-direction: column; gap: 2px; }
.banner-stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; }
.banner-stat-value { font-size: 16px; font-weight: 700; font-family: 'SF Mono', 'Cascadia Code', monospace; }

/* Error section */
.result-error-section { display: flex; flex-direction: column; gap: 10px; }
.error-block { background: var(--bg-hover); border: 1px solid var(--border-light); border-radius: 8px; padding: 12px; }
.error-block-title { font-size: 12px; font-weight: 700; color: var(--text-muted); margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
.error-pre { margin: 0; white-space: pre-wrap; word-break: break-word; font-size: 12px; line-height: 1.6; font-family: 'SF Mono', 'Cascadia Code', monospace; }
.error-text { color: var(--accent-red); }

/* Failed cases */
.result-cases-section { display: flex; flex-direction: column; gap: 10px; }
.cases-title { font-size: 14px; font-weight: 700; }
.case-card { border: 1px solid var(--border-light); border-radius: 8px; overflow: hidden; }
.case-header { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; background: var(--bg-hover); font-size: 13px; font-weight: 600; }
.case-body { padding: 12px; display: flex; flex-direction: column; gap: 10px; }
.case-field { display: flex; flex-direction: column; gap: 4px; }
.case-field-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; }
.case-field-value { margin: 0; white-space: pre-wrap; word-break: break-word; font-size: 12px; line-height: 1.5; font-family: 'SF Mono', 'Cascadia Code', monospace; background: var(--bg-card); border: 1px solid var(--border-light); border-radius: 6px; padding: 8px 10px; }
.case-compare { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.expected-label { color: var(--accent-green); }
.actual-label { color: var(--accent-red); }
.expected-value { border-color: var(--accent-green); background: #f0fdf4; }
.actual-value { border-color: var(--accent-red); background: #fef2f2; }
[data-theme="dark"] .expected-value { background: #052e16; border-color: #16a34a; }
[data-theme="dark"] .actual-value { background: #450a0a; border-color: #dc2626; }
.case-meta { display: flex; gap: 16px; padding: 8px 12px; border-top: 1px solid var(--border-light); font-size: 11px; color: var(--text-muted); }
@media (max-width: 600px) { .case-compare { grid-template-columns: 1fr; } }

/* AI section */
.result-ai-section { border: 1px solid var(--border-light); border-radius: 10px; padding: 14px; background: var(--bg-hover); }
.ai-section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.ai-section-title { display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 700; color: var(--accent-primary); }
.ai-loading { display: flex; align-items: center; gap: 8px; padding: 8px 0; color: var(--text-muted); font-size: 13px; }
.ai-content { font-size: 13px; line-height: 1.7; }
.ai-placeholder { font-size: 12px; color: var(--text-muted); }

/* Divider */
.divider { width: 4px; cursor: col-resize; background: var(--border-light); transition: background 0.15s; flex-shrink: 0; }
.divider:hover { background: var(--accent-primary); }

/* Editor area */
.editor-area { flex: 1; overflow: hidden; min-height: 200px; }

/* Test case panel */
.testcase-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-top: 1px solid #e8e8e8;
  position: relative;
}

/* Resize handle */
.panel-resize-handle {
  height: 5px;
  cursor: row-resize;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
  flex-shrink: 0;
}
.panel-resize-handle:hover { background: #1a73e8; }
.panel-resize-handle:hover .resize-dots { opacity: 1; }
.resize-dots {
  width: 24px; height: 3px;
  background: repeating-linear-gradient(90deg, #ccc 0px, #ccc 2px, transparent 2px, transparent 5px);
  opacity: 0.6;
  transition: opacity 0.15s;
}

/* Header */
.testcase-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 14px; height: 42px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  user-select: none;
  flex-shrink: 0;
}
.testcase-header-left { display: flex; align-items: center; flex: 1; min-width: 0; }
.testcase-header-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.testcase-tabs { display: flex; gap: 0; overflow-x: auto; }
.tc-tab {
  border: none; background: transparent;
  color: #666; padding: 10px 14px;
  font-size: 13px; cursor: pointer;
  white-space: nowrap; transition: all 0.12s;
  border-bottom: 2px solid transparent;
}
.tc-tab:hover { color: #333; background: #f5f5f5; }
.tc-tab.active { color: #1a1a1a; font-weight: 600; border-bottom-color: #1a73e8; }
.tc-tab.tc-add { font-size: 16px; font-weight: 400; padding: 10px 12px; color: #999; }
.tc-tab.tc-add:hover { color: #1a73e8; }
.collapse-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 4px;
  cursor: pointer; color: #999; transition: all 0.15s;
}
.collapse-btn:hover { background: #f0f0f0; color: #666; }

/* Body */
.testcase-body {
  flex: 1;
  padding: 12px 14px;
  overflow-y: auto;
  display: flex; flex-direction: column; gap: 10px;
  background: #fafafa;
}
.tc-section { display: flex; flex-direction: column; gap: 4px; }
.tc-label {
  font-size: 12px; color: #666; font-weight: 600;
  display: flex; align-items: center; gap: 8px;
  text-transform: uppercase; letter-spacing: 0.5px;
}
.tc-textarea {
  background: #fff; color: #1a1a1a;
  border: 1px solid #e0e0e0; border-radius: 6px;
  padding: 10px 12px;
  font-size: 13px; font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
  resize: none; outline: none; line-height: 1.6;
  width: 100%; box-sizing: border-box;
  transition: border-color 0.15s;
}
.tc-textarea:focus { border-color: #1a73e8; box-shadow: 0 0 0 2px rgba(26,115,232,0.1); }
.tc-textarea::placeholder { color: #bbb; }
.tc-toggle { border: none; background: transparent; color: #1a73e8; font-size: 12px; cursor: pointer; padding: 0; }
.tc-toggle:hover { text-decoration: underline; }

/* Run button */
.run-btn { font-weight: 600; }

/* Run result inline */
.tc-run-result { border-top: 1px solid #e8e8e8; padding-top: 10px; display: flex; flex-direction: column; gap: 8px; }
.tc-run-header { display: flex; align-items: center; justify-content: space-between; font-size: 13px; }
.tc-run-stats { font-size: 12px; color: #999; font-weight: 400; }
.tc-run-output { display: flex; flex-direction: column; gap: 4px; }
.tc-run-status { padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.badge-accepted { background: #dcfce7; color: #166534; }
.badge-error { background: #fef2f2; color: #991b1b; }
.badge-ce { background: #fdf4ff; color: #86198f; }
.badge-tle { background: #fff7ed; color: #9a3412; }
.badge-mle { background: #fef2f2; color: #991b1b; }
.badge-ole { background: #fff7ed; color: #9a3412; }
.badge-system { background: #fef2f2; color: #991b1b; }
.badge-pending { background: #f0f9ff; color: #0369a1; }
.badge-running { background: #fefce8; color: #a16207; }
.badge-default { background: #f3f4f6; color: #374151; }

/* Collapsed state */
.testcase-collapsed {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 8px 14px; cursor: pointer;
  background: #fafafa; border-top: 1px solid #e8e8e8;
  transition: background 0.15s;
  flex-shrink: 0;
}
.testcase-collapsed:hover { background: #f0f0f0; }
.collapsed-hint { font-size: 12px; color: #999; }
.expand-icon { color: #999; font-size: 14px; }
.tc-pre {
  margin: 0; white-space: pre-wrap; word-break: break-word;
  font-size: 12px; font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
  background: #1e1e1e; border: 1px solid #3c3c3c; border-radius: 6px;
  padding: 10px 12px; line-height: 1.5; color: #e5e5e5;
  max-height: 200px; overflow-y: auto;
}
.error-text { color: #f87171; }

/* Transitions */
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.slide-down-enter-active, .slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from, .slide-down-leave-to { max-height: 0; opacity: 0; overflow: hidden; }
.slide-down-enter-to, .slide-down-leave-from { max-height: 500px; opacity: 1; }
.slide-right-enter-active, .slide-right-leave-active { transition: all 0.3s ease; }
.slide-right-enter-from, .slide-right-leave-to { transform: translateX(100%); opacity: 0; }

/* AI Chat */
.ai-chat-fab { position: fixed; bottom: 20px; right: 20px; width: 44px; height: 44px; border-radius: 50%; background: var(--gradient-hero); color: #fff; display: flex; align-items: center; justify-content: center; cursor: pointer; box-shadow: var(--shadow-lg); z-index: 1001; transition: transform 0.15s; }
.ai-chat-fab:hover { transform: scale(1.08); }
.ai-chat-panel { position: fixed; bottom: 76px; right: 20px; width: 360px; height: 480px; background: var(--bg-card); border: 1px solid var(--border-light); border-radius: 12px; box-shadow: var(--shadow-xl); z-index: 1001; display: flex; flex-direction: column; overflow: hidden; }
.ai-chat-header { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; border-bottom: 1px solid var(--border-light); font-weight: 600; font-size: 14px; background: var(--bg-hover); }
.ai-chat-messages { flex: 1; overflow-y: auto; padding: 10px; display: flex; flex-direction: column; gap: 8px; }
.chat-msg { max-width: 85%; padding: 8px 12px; border-radius: 12px; font-size: 13px; line-height: 1.6; word-break: break-word; }
.chat-msg.user { align-self: flex-end; background: var(--accent-primary); color: #fff; border-bottom-right-radius: 4px; }
.chat-msg.assistant { align-self: flex-start; background: var(--bg-hover); color: var(--text-primary); border-bottom-left-radius: 4px; }
.ai-chat-input { padding: 10px; border-top: 1px solid var(--border-light); }
.chat-context-hint { margin-top: 4px; text-align: right; }

@media (max-width: 1200px) { .sample-grid { grid-template-columns: 1fr; } }
</style>
