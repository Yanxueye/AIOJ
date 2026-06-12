<template>
  <div class="knowledge-page page-container">
    <div class="page-header">
      <h2>知识图谱</h2>
      <p class="page-desc">基于 OI-Wiki 的算法知识点体系，点击节点查看相关题目</p>
    </div>

    <div class="knowledge-layout">
      <div class="graph-panel card">
        <div v-loading="loading" class="graph-container">
          <div ref="chartRef" class="chart-box" />
        </div>
        <div class="graph-controls">
          <el-select v-model="selectedCategory" placeholder="筛选分类" clearable size="small" @change="filterByCategory">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
        </div>
      </div>

      <div class="detail-panel">
        <div v-if="selectedKP" class="card kp-detail">
          <div class="kp-header">
            <span class="kp-icon" :style="{ background: selectedKP.color || 'var(--accent-primary)' }">
              {{ selectedKP.icon || selectedKP.name.charAt(0) }}
            </span>
            <div>
              <h3 class="kp-name">{{ selectedKP.name }}</h3>
              <span class="kp-category">{{ selectedKP.category }}</span>
            </div>
          </div>
          <p v-if="selectedKP.description" class="kp-desc">{{ selectedKP.description }}</p>
          <div v-if="selectedKP.ojWikiUrl" class="kp-link">
            <a :href="selectedKP.ojWikiUrl" target="_blank" rel="noopener">
              <el-icon><Link /></el-icon> OI-Wiki 参考
            </a>
          </div>
          <el-divider />
          <div class="kp-stats">
            <div class="kp-stat">
              <span class="stat-num">{{ kpProblemCount }}</span>
              <span class="stat-label">关联题目</span>
            </div>
            <div class="kp-stat">
              <span class="stat-num">{{ kpMastery }}%</span>
              <span class="stat-label">掌握度</span>
            </div>
          </div>
          <el-divider />
          <div class="kp-problems">
            <div class="section-label">相关题目</div>
            <div v-if="kpProblems.length" class="problem-list">
              <router-link
                v-for="p in kpProblems"
                :key="p.id"
                :to="`/problem/${p.id}`"
                class="problem-item"
              >
                <span class="problem-id">#{{ p.id }}</span>
                <span class="problem-title">{{ p.title }}</span>
                <el-tag :type="diffTagType(p.difficulty)" size="small" effect="plain">{{ p.difficulty }}</el-tag>
              </router-link>
            </div>
            <el-empty v-else description="暂无关联题目" :image-size="60" />
          </div>
        </div>
        <div v-else class="card kp-placeholder">
          <el-icon :size="48" style="color: var(--text-muted)"><DataBoard /></el-icon>
          <p>点击图谱中的节点查看知识点详情</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import * as echarts from 'echarts'
import http from '@/api/index'

const loading = ref(true)
const chartRef = ref(null)
const graphData = ref({ nodes: [], edges: [], counts: {}, mastery: {} })
const selectedCategory = ref('')
const selectedKP = ref(null)
const kpProblems = ref([])
const categories = ref([])

let chart = null

const CATEGORY_COLORS = {
  '动态规划': '#52c41a',
  '图论': '#3b82f6',
  '数据结构': '#f59e0b',
  '数学': '#ef4444',
  '字符串': '#8b5cf6',
  '搜索': '#13c2c2',
  '贪心': '#ec4899',
  '计算几何': '#f97316',
  '基础算法': '#6366f1',
  '位运算': '#84cc16'
}

const kpProblemCount = computed(() => {
  if (!selectedKP.value) return 0
  return graphData.value.counts[selectedKP.value.id] || 0
})

const kpMastery = computed(() => {
  if (!selectedKP.value) return 0
  return Math.round(graphData.value.mastery[selectedKP.value.id] || 0)
})

onMounted(async () => {
  try {
    const res = await http.get('/knowledge/graph')
    graphData.value = res.data || { nodes: [], edges: [], counts: {}, mastery: {} }
    categories.value = [...new Set(graphData.value.nodes.map(n => n.category))]
  } catch {
    graphData.value = { nodes: [], edges: [], counts: {}, mastery: {} }
  }
  loading.value = false
  await nextTick()
  initChart()
})

onUnmounted(() => {
  if (chart) {
    chart.dispose()
    chart = null
  }
})

function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  updateChart()

  chart.on('click', (params) => {
    if (params.dataType === 'node') {
      const node = graphData.value.nodes.find(n => n.id === params.data.id)
      if (node) {
        selectedKP.value = node
        loadProblems(node.id)
        showProblemNodes(node)
      }
    }
  })

  window.addEventListener('resize', () => chart?.resize())
}

function updateChart() {
  if (!chart) return

  const nodes = graphData.value.nodes.map(n => ({
    id: n.id,
    name: n.name,
    category: categories.value.indexOf(n.category),
    symbolSize: Math.max(20, Math.min(50, 15 + (graphData.value.counts[n.id] || 0) * 3)),
    itemStyle: {
      color: CATEGORY_COLORS[n.category] || '#52c41a',
      borderColor: '#fff',
      borderWidth: 2,
      shadowBlur: 8,
      shadowColor: 'rgba(0,0,0,0.1)'
    },
    label: {
      show: true,
      fontSize: 11
    }
  }))

  const links = graphData.value.edges.map(e => ({
    source: String(e.source),
    target: String(e.target),
    lineStyle: { color: '#c8d0be', width: 1 }
  }))

  const categoryObjs = categories.value.map(c => ({ name: c }))

  const style = getComputedStyle(document.documentElement)
  const textColor = style.getPropertyValue('--text-primary').trim() || '#1a2e1a'
  const mutedColor = style.getPropertyValue('--text-secondary').trim() || '#4a5d4a'

  chart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (params.dataType === 'node') {
          const count = graphData.value.counts[params.data.id] || 0
          const mastery = Math.round(graphData.value.mastery[params.data.id] || 0)
          return `<b>${params.name}</b><br/>关联题目: ${count}<br/>掌握度: ${mastery}%`
        }
        return ''
      }
    },
    legend: {
      data: categories.value,
      bottom: 0,
      textStyle: { color: mutedColor, fontSize: 11 }
    },
    series: [{
      type: 'graph',
      layout: 'force',
      data: nodes,
      links: links,
      categories: categoryObjs,
      roam: true,
      draggable: true,
      force: {
        repulsion: 300,
        edgeLength: [80, 160],
        gravity: 0.1
      },
      emphasis: {
        focus: 'adjacency',
        lineStyle: { width: 3 }
      },
      label: {
        position: 'bottom',
        color: textColor
      }
    }]
  }, true)
}

function filterByCategory(cat) {
  if (!chart) return
  if (!cat) {
    updateChart()
    return
  }
  const filteredNodes = graphData.value.nodes.filter(n => n.category === cat)
  const filteredIDs = new Set(filteredNodes.map(n => n.id))
  const filteredEdges = graphData.value.edges.filter(e => filteredIDs.has(e.source) && filteredIDs.has(e.target))

  const nodes = filteredNodes.map(n => ({
    id: n.id,
    name: n.name,
    category: categories.value.indexOf(n.category),
    symbolSize: Math.max(20, Math.min(50, 15 + (graphData.value.counts[n.id] || 0) * 3)),
    itemStyle: { color: CATEGORY_COLORS[n.category] || '#52c41a', borderColor: '#fff', borderWidth: 2 },
    label: { show: true, fontSize: 11 }
  }))

  const links = filteredEdges.map(e => ({
    source: String(e.source),
    target: String(e.target),
    lineStyle: { color: '#c8d0be', width: 1 }
  }))

  chart.setOption({
    series: [{
      data: nodes,
      links: links
    }]
  }, true)
}

async function loadProblems(kpID) {
  try {
    const res = await http.get(`/knowledge/${kpID}/problems`)
    kpProblems.value = res.data?.items || []
  } catch {
    kpProblems.value = []
  }
}

function showProblemNodes(kpNode) {
  if (!chart || !kpProblems.value.length) return
  removeProblemNodes()

  const kpNodeData = {
    id: kpNode.id,
    name: kpNode.name,
    category: categories.value.indexOf(kpNode.category),
    symbolSize: Math.max(20, Math.min(50, 15 + (graphData.value.counts[kpNode.id] || 0) * 3)),
    itemStyle: { color: CATEGORY_COLORS[kpNode.category] || '#52c41a', borderColor: '#fff', borderWidth: 2 },
    label: { show: true, fontSize: 11 }
  }

  const problemNodes = kpProblems.value.slice(0, 8).map(p => ({
    id: `problem-${p.id}`,
    name: `#${p.id}`,
    symbolSize: 16,
    category: -1,
    itemStyle: {
      color: p.difficulty === '简单' ? '#52c41a' : p.difficulty === '中等' ? '#f59e0b' : '#ef4444',
      borderColor: '#fff',
      borderWidth: 2
    },
    label: { show: true, fontSize: 10, position: 'right' }
  }))

  const problemLinks = kpProblems.value.slice(0, 8).map(p => ({
    source: String(kpNode.id),
    target: `problem-${p.id}`,
    lineStyle: { color: '#aaa', width: 1, type: 'dashed' }
  }))

  chart.setOption({
    series: [{
      data: [kpNodeData, ...problemNodes],
      links: problemLinks
    }]
  })

  chart.on('click', 'series.graph', (params) => {
    if (params.dataType === 'node' && String(params.data.id).startsWith('problem-')) {
      const pid = String(params.data.id).replace('problem-', '')
      window.open(`/problem/${pid}`, '_blank')
    }
  })
}

function removeProblemNodes() {
  if (!chart) return
  const option = chart.getOption()
  if (!option.series?.[0]) return
  const currentData = option.series[0].data || []
  const currentLinks = option.series[0].links || []
  const filteredData = currentData.filter(n => !String(n.id).startsWith('problem-'))
  const filteredLinks = currentLinks.filter(l => !String(l.target).startsWith('problem-'))
  chart.setOption({
    series: [{ data: filteredData, links: filteredLinks }]
  })
}

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
}
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}

.knowledge-layout {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 20px;
  min-height: 600px;
}

.graph-panel {
  display: flex;
  flex-direction: column;
}

.graph-container {
  flex: 1;
  min-height: 500px;
}

.chart-box {
  width: 100%;
  height: 100%;
  min-height: 500px;
}

.graph-controls {
  padding: 12px 0 0;
  display: flex;
  gap: 12px;
}

.detail-panel {
  display: flex;
  flex-direction: column;
}

.kp-detail {
  flex: 1;
  background: var(--gradient-card);
}

.kp-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.kp-icon {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  flex-shrink: 0;
}

.kp-name {
  font-size: 18px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.kp-category {
  font-size: 12px;
  color: var(--text-muted);
}

.kp-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.7;
}

.kp-link a {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--accent-primary);
  font-size: 13px;
  font-weight: 500;
}

.kp-link a:hover {
  text-decoration: underline;
}

.kp-stats {
  display: flex;
  gap: 24px;
}

.kp-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-num {
  font-size: 22px;
  font-weight: 800;
  color: var(--accent-gold);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

.section-label {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 10px;
  color: var(--text-primary);
}

.problem-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.problem-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-light);
  background: var(--bg-warm);
  text-decoration: none;
  color: inherit;
  transition: all var(--transition-fast);
}

.problem-item:hover {
  border-color: var(--accent-primary);
  background: var(--accent-primary-bg);
}

.problem-id {
  font-family: var(--font-mono);
  color: var(--text-muted);
  min-width: 40px;
  font-size: 12px;
}

.problem-title {
  flex: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.kp-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 300px;
  color: var(--text-muted);
  font-size: 14px;
}

@media (max-width: 960px) {
  .knowledge-layout {
    grid-template-columns: 1fr;
  }
}
</style>
