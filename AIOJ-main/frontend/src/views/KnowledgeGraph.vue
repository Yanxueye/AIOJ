<template>
  <div class="knowledge-page page-container">
    <div class="page-header">
      <div class="page-header-top">
        <div>
          <h2>知识图谱</h2>
          <p class="page-desc">基于 OI-Wiki 的完整算法知识体系，点击节点查看推荐题目与学习建议</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedCategory" placeholder="筛选分类" clearable size="default" style="width: 160px" @change="handleCategoryFilter">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
          <el-button size="default" type="primary" plain :loading="aiLoading" @click="generateAIGraph">
            <el-icon><MagicStick /></el-icon> AI 分析薄弱点
          </el-button>
          <el-button size="default" type="success" plain :loading="planLoading" @click="generateStudyPlan">
            <el-icon><MagicStick /></el-icon> AI 创建题单
          </el-button>
        </div>
      </div>
    </div>

    <div class="knowledge-layout">
      <div class="graph-panel card">
        <div v-loading="loading" class="graph-container">
          <div ref="chartRef" class="chart-box" />
        </div>
        <div class="graph-legend">
          <div class="legend-section">
            <span class="legend-title">掌握度</span>
            <div class="legend-items">
              <span class="legend-item"><i class="legend-dot" style="background:#c0c4cc" /> 未学习</span>
              <span class="legend-item"><i class="legend-dot" style="background:#f56c6c" /> 学习中</span>
              <span class="legend-item"><i class="legend-dot" style="background:#e6a23c" /> 熟悉</span>
              <span class="legend-item"><i class="legend-dot" style="background:#409eff" /> 精通</span>
              <span class="legend-item"><i class="legend-dot" style="background:#67c23a" /> 掌握</span>
            </div>
          </div>
          <div class="legend-section">
            <span class="legend-title">距离</span>
            <span class="legend-hint">中心 = 基础，外围 = 进阶</span>
          </div>
        </div>
      </div>

      <div class="detail-panel">
        <div v-if="selectedKP" class="card kp-detail">
          <div class="kp-header">
            <span class="kp-icon" :style="{ background: selectedKP.color || getCategoryColor(selectedKP.category) }">
              {{ selectedKP.name.charAt(0) }}
            </span>
            <div class="kp-header-text">
              <h3 class="kp-name">{{ displayName(selectedKP.name) }}</h3>
              <el-tag size="small" :style="{ background: getCategoryColor(selectedKP.category) + '18', color: getCategoryColor(selectedKP.category), border: 'none' }">
                {{ selectedKP.category }}
              </el-tag>
            </div>
            <el-button class="kp-close" :icon="Close" text circle size="small" @click="clearSelection" />
          </div>
          <p v-if="selectedKP.description" class="kp-desc">{{ selectedKP.description }}</p>
          <div v-if="selectedKP.ojWikiUrl" class="kp-link">
            <a :href="selectedKP.ojWikiUrl" target="_blank" rel="noopener"><el-icon><Link /></el-icon> OI-Wiki 参考</a>
          </div>
          <el-divider />
          <div class="kp-stats">
            <div class="kp-stat">
              <span class="stat-num">{{ kpProblemCount }}</span>
              <span class="stat-label">关联题目</span>
            </div>
            <div class="kp-stat">
              <span class="stat-num" :style="{ color: masteryColor(kpMastery) }">{{ kpMastery }}%</span>
              <span class="stat-label">掌握度</span>
            </div>
          </div>
          <el-progress :percentage="kpMastery" :color="masteryColor(kpMastery)" :stroke-width="10" style="margin: 12px 0" />
          <el-divider />
          <div class="kp-problems">
            <div class="section-label">推荐题目</div>
            <div v-if="kpProblemsLoading" class="problems-loading"><el-skeleton :rows="3" animated /></div>
            <div v-else-if="kpUntried.length" class="problem-list">
              <router-link v-for="p in kpUntried" :key="p.id" :to="`/problem/${p.id}`" class="problem-item problem-untried">
                <span class="problem-id">#{{ p.id }}</span>
                <span class="problem-title">{{ p.title }}</span>
                <el-tag :type="diffTagType(p.difficulty)" size="small" effect="plain">{{ p.difficulty }}</el-tag>
              </router-link>
            </div>
            <el-empty v-else description="暂无关联题目" :image-size="48" />
          </div>
          <div v-if="aiSuggestions.length" class="kp-suggestions">
            <el-divider />
            <div class="section-label"><el-icon><MagicStick /></el-icon> AI 学习建议</div>
            <ul class="suggestion-list"><li v-for="(s, i) in aiSuggestions" :key="i">{{ s }}</li></ul>
          </div>
        </div>

        <div v-else-if="aiAnalysisActive" class="card overview-panel">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px">
            <h3 class="overview-title" style="margin:0">AI 分析</h3>
            <el-button size="small" text @click="stopPulse(); aiAnalysisActive = false; highlightedNodes.value = new Map(); renderChart()">× 关闭</el-button>
          </div>
          <!-- 掌握较好 -->
          <div v-if="aiStrengths.length" class="section-label" style="color:#67c23a">🟢 掌握较好</div>
          <div v-if="aiStrengths.length" style="display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px">
            <el-tag v-for="t in aiStrengths" :key="t" size="small" effect="plain" type="success">{{ t }}</el-tag>
          </div>
          <!-- 需要加强 -->
          <div v-if="aiWeaknesses.length" class="section-label" style="color:#f56c6c">🔴 需要加强</div>
          <div v-if="aiWeaknesses.length" style="display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px">
            <el-tag v-for="t in aiWeaknesses" :key="t" size="small" effect="plain" type="danger">{{ t }}</el-tag>
          </div>
          <el-divider />
          <div v-if="aiSuggestions.length" class="section-label">学习建议</div>
          <ul v-if="aiSuggestions.length" class="suggestion-list"><li v-for="(s,i) in aiSuggestions" :key="i">{{ s }}</li></ul>
        </div>

        <div v-else class="card overview-panel">
          <h3 class="overview-title">知识体系总览</h3>
          <div class="overview-stats">
            <div class="overview-stat"><span class="stat-num">{{ totalKPCount }}</span><span class="stat-label">知识点</span></div>
            <div class="overview-stat"><span class="stat-num">{{ totalProblemCount }}</span><span class="stat-label">关联题目</span></div>
            <div class="overview-stat"><span class="stat-num">{{ learnedCount }}</span><span class="stat-label">已学习</span></div>
            <div class="overview-stat"><span class="stat-num" :style="{ color: masteryColor(avgMastery) }">{{ avgMastery }}%</span><span class="stat-label">平均掌握</span></div>
          </div>
          <el-divider />
          <div class="section-label">分类完成度</div>
          <div class="category-progress-list">
            <div v-for="cat in categoryStats" :key="cat.name" class="category-progress-item">
              <div class="cat-progress-header">
                <span class="cat-dot" :style="{ background: cat.color }" /><span class="cat-name">{{ cat.name }}</span><span class="cat-count">{{ cat.learned }}/{{ cat.total }}</span>
              </div>
              <el-progress :percentage="cat.total > 0 ? Math.round(cat.learned / cat.total * 100) : 0" :stroke-width="6" :show-text="false" :color="cat.color" />
            </div>
          </div>
          <el-divider />
          <p class="overview-hint"><el-icon><InfoFilled /></el-icon> 点击左侧图谱中的节点查看知识点详情与推荐题目</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import * as echarts from 'echarts'
import http from '@/api/index'
import { aiApi } from '@/api/ai'
import { ElMessage } from 'element-plus'
import { Close, InfoFilled, MagicStick, Link } from '@element-plus/icons-vue'

const loading = ref(true)
const aiLoading = ref(false)
const planLoading = ref(false)
const chartRef = ref(null)
const graphData = ref({ nodes: [], edges: [], counts: {}, mastery: {} })
const selectedCategory = ref('')
const selectedKP = ref(null)
const kpUntried = ref([])
const kpTried = ref([])
const kpProblemsLoading = ref(false)
const categories = ref([])
const aiSuggestions = ref([])
const aiAnalysisActive = ref(false)
const aiStrengths = ref([])   // 掌握较好的标签
const aiWeaknesses = ref([])  // 需要加强的标签
const highlightedNodes = ref(new Map())
const pulseShadow = ref(24)  // pulsing glow
let chart = null
let pulseTimer = null
let resizeObserver = null

const CATEGORY_COLORS = {
  '基础算法': '#6366f1', '数据结构': '#f59e0b', '动态规划': '#3b82f6', '图论': '#10b981',
  '数学': '#8b5cf6', '字符串': '#ef4444', '搜索': '#06b6d4', '贪心': '#f97316',
  '计算几何': '#ec4899', '位运算': '#14b8a6'
}
function getCategoryColor(cat) { return CATEGORY_COLORS[cat] || '#6b7280' }
function displayName(name) { return name ? name.replace(/（分类）$/, '') : '' }
function masteryColor(l) { if (l>=80) return '#67c23a'; if (l>=60) return '#409eff'; if (l>=40) return '#e6a23c'; if (l>0) return '#f56c6c'; return '#c0c4cc' }
function masteryLevelName(l) { if (l>=80) return '掌握'; if (l>=60) return '精通'; if (l>=40) return '熟悉'; if (l>0) return '学习中'; return '未学习' }
function diffTagType(d) { return d==='简单'?'success':d==='中等'?'warning':'danger' }

const kpProblemCount = computed(() => selectedKP.value ? graphData.value.counts[selectedKP.value.id] || 0 : 0)
const kpMastery = computed(() => selectedKP.value ? Math.round(graphData.value.mastery[selectedKP.value.id] || 0) : 0)
const totalKPCount = computed(() => graphData.value.nodes.length)
const totalProblemCount = computed(() => Object.values(graphData.value.counts).reduce((a, b) => a + b, 0))
const learnedCount = computed(() => Object.values(graphData.value.mastery).filter(v => v > 0).length)
const avgMastery = computed(() => {
  const vals = Object.values(graphData.value.mastery).filter(v => v > 0)
  return vals.length ? Math.round(vals.reduce((a,b)=>a+b,0)/vals.length) : 0
})
const categoryStats = computed(() => {
  const stats = {}
  graphData.value.nodes.forEach(n => {
    if (!stats[n.category]) stats[n.category] = { name: n.category, color: getCategoryColor(n.category), total: 0, learned: 0 }
    stats[n.category].total++
    if ((graphData.value.mastery[n.id]||0) > 0) stats[n.category].learned++
  })
  return Object.values(stats)
})

// ── 父节点 ID 集合（有子节点的 ID）──────────────────
let parentIds = new Set()

// ── 子树题目数 ──────────────────────────────────────
function subCount(nodeId) {
  const { nodes, counts } = graphData.value
  let n = counts[nodeId] || 0
  nodes.filter(x => x.parentId === nodeId).forEach(c => { n += subCount(c.id) })
  return n
}

// ── 构建树 ──────────────────────────────────────────
function buildTree(filterCategory) {
  const { nodes, counts, mastery } = graphData.value
  const list = filterCategory ? nodes.filter(n => n.category === filterCategory) : nodes
  const byParent = {}
  list.forEach(n => { const p = n.parentId || 0; (byParent[p] = byParent[p] || []).push(n) })

  // 记录父节点
  parentIds = new Set()
  Object.keys(byParent).forEach(k => { if (k !== '0') parentIds.add(Number(k)) })

  function walk(kp) {
    const c = counts[kp.id] || 0
    const sc = subCount(kp.id)
    const m = mastery[kp.id] || 0
    const hasPractice = m > 0
    const color = hasPractice ? getCategoryColor(kp.category) : '#c0c4cc'
    const kids = (byParent[kp.id] || []).map(walk)
    const leaf = kids.length === 0
    const node = {
      name: displayName(kp.name), value: c, knowledgeId: kp.id,
      category: kp.category, description: kp.description, ojWikiUrl: kp.ojWikiUrl,
      color, mastery: m, subCount: sc,
      itemStyle: { color, borderColor: highlightedNodes.value.get(kp.id) === 'weakness' ? '#f56c6c' : highlightedNodes.value.get(kp.id) === 'strength' ? '#67c23a' : hasPractice ? masteryColor(m) : '#dcdfe6', borderWidth: highlightedNodes.value.has(kp.id) ? 5 : hasPractice ? 3 : 1, opacity: 1, shadowBlur: highlightedNodes.value.has(kp.id) ? (pulseShadow.value || 24) : 0, shadowColor: highlightedNodes.value.get(kp.id) === 'weakness' ? '#f56c6c' : highlightedNodes.value.get(kp.id) === 'strength' ? '#67c23a' : 'transparent' },
      lineStyle: { width: Math.max(1, Math.min(6, 1 + sc * 0.3)), color: hasPractice ? getCategoryColor(kp.category)+'60' : '#d1d5db' },
      symbolSize: leaf ? Math.max(12, Math.min(24, 8+c*1.5)) : Math.max(18, Math.min(32, 12+c*1.5)),
      label: { fontSize: leaf ? 11 : 13, fontWeight: leaf ? 400 : 700, color: hasPractice ? '#303133' : '#909399' },
    }
    if (kids.length) node.children = kids
    return node
  }

  const roots = (byParent[0] || []).map(walk)
  if (filterCategory && roots.length === 1) return roots
  return [{ name: '算法知识', symbol: 'none', label: { show: false }, itemStyle: { opacity: 0 }, lineStyle: { opacity: 0 }, children: roots }]
}

// ── 渲染 ────────────────────────────────────────────
function renderChart() {
  if (!chart) return
  const opts = {
    tooltip: {
      trigger: 'item', confine: true,
      backgroundColor: 'rgba(30,30,30,0.95)', borderColor: '#3c3c3c',
      textStyle: { color: '#e5e5e5', fontSize: 12 },
      formatter(p) {
        const d = p.data; if (!d?.knowledgeId) return d?.name || ''
        const m = d.mastery || 0; const isP = parentIds.has(d.knowledgeId)
        const sub = d.subCount > (d.value||0) ? ` · 子树共 ${d.subCount} 题` : ''
        return `<div style="font-weight:700">${d.name}</div><div style="color:#888;font-size:11px">${d.category}</div><div style="margin-top:6px">关联题目: <b>${d.value||0}</b>${sub}</div><div>掌握度: <b style="color:${masteryColor(m)}">${Math.round(m)}% (${masteryLevelName(m)})</b></div><div style="color:#409eff;font-size:11px;margin-top:4px">${isP?'点击折叠/展开子树':'点击查看推荐题目'}</div>`
      }
    },
    series: [{
      type: 'tree', data: buildTree(selectedCategory.value), layout: 'radial',
      symbol: 'circle', roam: true, expandAndCollapse: true,
      nodeClick: true, cursor: 'pointer',
      animationDuration: 400, animationDurationUpdate: 300,
      lineStyle: { curveness: 0.5 },
      label: { position: 'right', distance: 8 },
      labelLayout: { hideOverlap: true },
      leaves: { label: { position: 'right', distance: 8 } },
      emphasis: { focus: 'ancestor', itemStyle: { shadowBlur: 20, shadowColor: 'rgba(0,0,0,0.3)' } },
    }]
  }
  chart.setOption(opts, true)
}

// ── 初始化 ──────────────────────────────────────────
function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  renderChart()

  // 点击事件：叶子节点加载题目
  chart.on('click', (params) => {
    if (!params || !params.data || !params.data.knowledgeId) return
    const d = params.data
    if (parentIds.has(d.knowledgeId)) return // 非叶子，ECharts 自己处理折叠

    // 叶子节点：加载题目
    const kp = graphData.value.nodes.find(n => n.id === d.knowledgeId)
    if (kp) {
      selectedKP.value = { ...kp, color: kp.color || getCategoryColor(kp.category) }
      aiSuggestions.value = []
      loadProblems(kp.id)
    }
  })

  resizeObserver = new ResizeObserver(() => chart?.resize())
  resizeObserver.observe(chartRef.value)
}

// ── 其他 ────────────────────────────────────────────
function handleCategoryFilter() { clearSelection(); renderChart() }
function clearSelection() { selectedKP.value = null; kpUntried.value = []; kpTried.value = []; aiSuggestions.value = [] }

async function loadProblems(id) {
  kpProblemsLoading.value = true
  try {
    const r = await http.get(`/knowledge/${id}/problems`)
    const d = r.data || {}
    kpUntried.value = d.untried || []
    kpTried.value = d.tried || []
  } catch { kpUntried.value = []; kpTried.value = [] }
  finally { kpProblemsLoading.value = false }
}

async function generateAIGraph() {
  aiLoading.value = true
  try {
    const r = await aiApi.buildKnowledgeGraph({ scope: 'recent' })
    const d = r.data
    if (d?.nodes?.length || d?.suggestions?.length) {
      const strengths = [], weaknesses = [], highlightMap = new Map()
      d.nodes.forEach(n => {
        const name = n.label || n.id
        const kp = graphData.value.nodes.find(k => k.name === name)
        if (!kp) return
        if (n.mastery === 'mastered' || n.mastery === 'proficient' || n.mastery === 'familiar') {
          strengths.push(name)
          highlightMap.set(kp.id, 'strength')
        } else {
          weaknesses.push(name)
          highlightMap.set(kp.id, 'weakness')
        }
      })
      aiStrengths.value = strengths
      aiWeaknesses.value = weaknesses
      aiSuggestions.value = d.suggestions || []
      aiAnalysisActive.value = true
      selectedKP.value = null
      kpUntried.value = []; kpTried.value = []
      highlightedNodes.value = highlightMap
      renderChart()
      startPulse()
      ElMessage.success('AI 分析完成')
    } else { ElMessage.info('AI 未返回分析数据') }
  } catch { ElMessage.error('AI 分析失败') }
  finally { aiLoading.value = false }
}

async function generateStudyPlan() {
  planLoading.value = true
  try {
    const r = await aiApi.createStudyPlan()
    const d = r.data
    if (d?.id) {
      ElMessage.success(`AI 题单"${d.title}"创建成功，${d.problemCount} 道题`)
    } else {
      ElMessage.info('AI 未能创建题单')
    }
  } catch { ElMessage.error('AI 创建题单失败') }
  finally { planLoading.value = false }
}

onMounted(async () => {
  try {
    const r = await http.get('/knowledge/graph')
    graphData.value = r.data || { nodes: [], edges: [], counts: {}, mastery: {} }
    categories.value = [...new Set(graphData.value.nodes.map(n => n.category))]
  } catch { graphData.value = { nodes: [], edges: [], counts: {}, mastery: {} } }
  loading.value = false
  await nextTick()
  initChart()
})

function startPulse() {
  stopPulse()
  let up = false
  pulseTimer = setInterval(() => {
    pulseShadow.value = up ? 30 : 18; up = !up
    renderChart()
  }, 600)
}
function stopPulse() {
  if (pulseTimer) { clearInterval(pulseTimer); pulseTimer = null }
  pulseShadow.value = 24
}

onUnmounted(() => {
  stopPulse()
  chart?.dispose(); chart = null
  resizeObserver?.disconnect(); resizeObserver = null
})
</script>

<style scoped>
.page-header { margin-bottom: 16px }
.page-header-top { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; flex-wrap: wrap }
.page-header h2 {
  font-family: 'JetBrains Mono', 'Cascadia Code', monospace;
  font-size: 26px; font-weight: 700; letter-spacing: -0.03em; margin: 0;
  background: linear-gradient(135deg, var(--text-primary) 30%, var(--accent-gold, #e6a23c) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
}
.page-desc { font-size: 13px; color: var(--text-muted); margin: 4px 0 0; font-family: var(--font-mono) }
.header-actions { display: flex; gap: 10px; align-items: center; flex-shrink: 0 }
.knowledge-layout { display: grid; grid-template-columns: 1fr 360px; gap: 16px; min-height: calc(100vh - 200px) }
.graph-panel { display: flex; flex-direction: column; min-height: 0 }
.graph-container { flex: 1; min-height: 700px }
.chart-box { width: 100%; height: 100%; min-height: 700px }
.graph-legend { display: flex; align-items: center; gap: 24px; padding: 10px 16px; border-top: 1px solid var(--border-light); flex-shrink: 0; flex-wrap: wrap }
.legend-section { display: flex; align-items: center; gap: 10px }
.legend-title { font-size: 11px; font-weight: 600; color: var(--text-muted); white-space: nowrap }
.legend-items { display: flex; gap: 12px; flex-wrap: wrap }
.legend-item { display: flex; align-items: center; gap: 4px; font-size: 11px; color: var(--text-secondary) }
.legend-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0 }
.legend-hint { font-size: 11px; color: var(--text-muted) }
.detail-panel { display: flex; flex-direction: column; min-height: 0 }
.kp-detail { flex: 1; overflow-y: auto; padding: 20px }
.kp-header { display: flex; align-items: flex-start; gap: 12px; margin-bottom: 12px }
.kp-icon { width: 44px; height: 44px; border-radius: 10px; display: flex; align-items: center; justify-content: center; color: #fff; font-weight: 700; font-size: 18px; flex-shrink: 0 }
.kp-header-text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px }
.kp-name { font-size: 18px; font-weight: 800; color: var(--text-primary); letter-spacing: -0.02em; margin: 0 }
.kp-close { flex-shrink: 0; margin-top: 2px }
.kp-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.7; margin: 0 0 8px }
.kp-link a { display: inline-flex; align-items: center; gap: 4px; color: var(--accent-primary); font-size: 13px; font-weight: 500; text-decoration: none }
.kp-link a:hover { text-decoration: underline }
.kp-stats { display: flex; gap: 24px }
.kp-stat { display: flex; flex-direction: column; align-items: center }
.stat-num { font-size: 22px; font-weight: 800; color: var(--accent-gold) }
.stat-label { font-size: 12px; color: var(--text-muted) }
.section-label { font-size: 13px; font-weight: 700; margin-bottom: 10px; color: var(--text-primary); display: flex; align-items: center; gap: 6px }
.problems-sections { display: flex; flex-direction: column; gap: 4px }
.section-subtitle { font-size: 12px; font-weight: 600; color: #6366f1; margin-bottom: 4px; margin-top: 2px }
.section-subtitle + .section-subtitle { color: var(--text-muted) }
.problem-items { display: flex; flex-direction: column; gap: 6px }
.problem-list { display: flex; flex-direction: column; gap: 6px; max-height: 320px; overflow-y: auto }
.problem-item { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-radius: 6px; border: 1px solid var(--border-light); background: var(--bg-warm, var(--bg-card)); text-decoration: none; color: inherit; transition: all 0.15s }
.problem-item:hover { border-color: var(--accent-primary); background: var(--accent-primary-bg, rgba(99, 102, 241, 0.05)) }
.problem-untried { border-left: 3px solid var(--accent-primary, #6366f1) }
.problem-id { font-family: monospace; color: var(--text-muted); min-width: 44px; font-size: 12px }
.problem-title { flex: 1; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px }
.problems-loading { padding: 8px 0 }
.overview-panel { padding: 20px }
.overview-title { font-size: 17px; font-weight: 800; color: var(--text-primary); margin: 0 0 16px }
.overview-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 12px }
.overview-stat { display: flex; flex-direction: column; align-items: center; padding: 12px 8px; border-radius: 8px; background: var(--bg-warm, var(--bg-card)); border: 1px solid var(--border-light) }
.category-progress-list { display: flex; flex-direction: column; gap: 10px }
.category-progress-item { display: flex; flex-direction: column; gap: 4px }
.cat-progress-header { display: flex; align-items: center; gap: 8px; font-size: 12px }
.cat-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0 }
.cat-name { flex: 1; color: var(--text-primary); font-weight: 500 }
.cat-count { color: var(--text-muted); font-family: monospace; font-size: 11px }
.overview-hint { font-size: 13px; color: var(--text-muted); display: flex; align-items: center; gap: 6px; margin: 0 }
.suggestion-list { padding-left: 18px; margin: 0; font-size: 13px; color: var(--text-secondary); line-height: 1.8 }
.suggestion-list li { margin-bottom: 4px }
@media (max-width: 960px) {
  .knowledge-layout { grid-template-columns: 1fr }
  .detail-panel { order: -1 }
  .page-header-top { flex-direction: column }
}
</style>
