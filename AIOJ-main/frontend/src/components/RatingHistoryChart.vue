<template>
  <div class="card section-card">
    <div class="section-title">Rating 变化曲线</div>
    <div v-if="loading" class="chart-loading">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
    </div>
    <div v-else-if="history.length === 0" class="chart-empty">
      <el-empty description="暂无 Rating 变化记录" :image-size="80" />
    </div>
    <div v-else ref="chartRef" class="chart-container" />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const props = defineProps({
  history: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false }
})

const chartRef = ref(null)
let chart = null

function renderChart() {
  if (!chartRef.value || props.history.length === 0) return

  if (chart) chart.dispose()
  chart = echarts.init(chartRef.value)

  const sorted = [...props.history].sort((a, b) => new Date(a.createdAt) - new Date(b.createdAt))
  const dates = sorted.map(h => {
    const d = new Date(h.createdAt)
    return `${d.getMonth() + 1}/${d.getDate()}`
  })
  const ratings = sorted.map(h => h.newRating)

  const minRating = Math.max(0, Math.min(...ratings) - 100)
  const maxRating = Math.max(...ratings) + 100

  // Color based on rating range
  const getColor = (r) => {
    if (r >= 2400) return '#ff0000'  // red
    if (r >= 2100) return '#ff8c00'  // orange
    if (r >= 1900) return '#a020f0'  // purple
    if (r >= 1600) return '#0000ff'  // blue
    if (r >= 1400) return '#03a89e'  // cyan
    if (r >= 1200) return '#008000'  // green
    return '#808080'                 // gray
  }

  const option = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1e1e1e',
      borderColor: '#3c3c3c',
      textStyle: { color: '#e5e5e5', fontSize: 12 },
      formatter: (params) => {
        const p = params[0]
        const h = sorted[p.dataIndex]
        const sign = h.delta >= 0 ? '+' : ''
        const color = h.delta >= 0 ? '#22c55e' : '#ef4444'
        return `<div style="font-weight:600">${p.name}</div>
          <div>Rating: <b>${h.newRating}</b></div>
          <div style="color:${color}">${sign}${h.delta}</div>
          <div style="color:#888;font-size:11px">#${h.problemId} · ${h.reason === 'submit' ? '提交' : h.reason}</div>`
      }
    },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { lineStyle: { color: '#3c3c3c' } },
      axisLabel: { color: '#888', fontSize: 11 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      min: minRating,
      max: maxRating,
      splitLine: { lineStyle: { color: '#2a2a2a' } },
      axisLine: { show: false },
      axisLabel: { color: '#888', fontSize: 11 }
    },
    series: [{
      type: 'line',
      data: ratings,
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 2.5, color: '#3b82f6' },
      itemStyle: {
        color: (params) => getColor(params.value),
        borderWidth: 2,
        borderColor: '#1e1e1e'
      },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(59,130,246,0.25)' },
          { offset: 1, color: 'rgba(59,130,246,0.02)' }
        ])
      },
      emphasis: {
        itemStyle: { shadowBlur: 10, shadowColor: 'rgba(59,130,246,0.5)' }
      }
    }]
  }

  chart.setOption(option)
}

function handleResize() {
  chart?.resize()
}

onMounted(() => {
  nextTick(renderChart)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
})

watch(() => props.history, () => nextTick(renderChart), { deep: true })
</script>

<style scoped>
.chart-container {
  width: 100%;
  height: 240px;
}

.chart-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 240px;
  color: var(--text-muted);
}

.chart-empty {
  height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
