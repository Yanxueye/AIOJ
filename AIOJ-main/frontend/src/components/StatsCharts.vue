<template>
  <div class="stats-charts">
    <div class="chart-row">
      <div class="chart-card card">
        <div class="card-title">难度分布</div>
        <div v-if="!hasDifficultyData" class="empty-state">暂无做题数据</div>
        <v-chart v-else :option="difficultyOption" autoresize style="height: 280px" />
      </div>
      <div class="chart-card card">
        <div class="card-title">算法分类统计</div>
        <div v-if="!hasAlgorithmData" class="empty-state">暂无做题数据</div>
        <v-chart v-else :option="algorithmOption" autoresize style="height: 280px" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const props = defineProps({
  difficultyData: { type: Object, default: () => ({}) },
  algorithmData: { type: Object, default: () => ({}) }
})

const DIFF_COLORS = { '简单': '#22c55e', '中等': '#f59e0b', '困难': '#ef4444' }

const hasDifficultyData = computed(() => Object.keys(props.difficultyData).length > 0)
const hasAlgorithmData = computed(() => Object.keys(props.algorithmData).length > 0)

const difficultyOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 题 ({d}%)' },
  legend: { bottom: 0, textStyle: { fontSize: 12 } },
  series: [{
    type: 'pie',
    radius: ['40%', '65%'],
    center: ['50%', '45%'],
    avoidLabelOverlap: true,
    itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
    label: { show: true, formatter: '{b}\n{c}题' },
    data: Object.entries(props.difficultyData).map(([name, value]) => ({
      name, value, itemStyle: { color: DIFF_COLORS[name] || '#409eff' }
    }))
  }]
}))

const algorithmOption = computed(() => {
  const entries = Object.entries(props.algorithmData).sort((a, b) => b[1] - a[1])
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 80, right: 20, top: 10, bottom: 30 },
    xAxis: { type: 'value', minInterval: 1 },
    yAxis: {
      type: 'category',
      data: entries.map(e => e[0]),
      inverse: true,
      axisLabel: { fontSize: 12 }
    },
    series: [{
      type: 'bar',
      data: entries.map(e => e[1]),
      barWidth: 18,
      itemStyle: {
        borderRadius: [0, 4, 4, 0],
        color: { type: 'linear', x: 0, y: 0, x2: 1, y2: 0,
          colorStops: [
            { offset: 0, color: '#52c41a' },
            { offset: 1, color: '#389e0d' }
          ]
        }
      },
      label: { show: true, position: 'right', formatter: '{c} 题', fontSize: 11 }
    }]
  }
})
</script>

<style scoped>
.chart-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}
.chart-card {
  min-height: 340px;
}
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 280px;
  color: #909399;
  font-size: 14px;
}
@media (max-width: 768px) {
  .chart-row {
    grid-template-columns: 1fr;
  }
}
</style>
