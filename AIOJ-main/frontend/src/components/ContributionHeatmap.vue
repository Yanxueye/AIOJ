<template>
  <div class="heatmap-container">
    <div class="heatmap-header">
      <span class="heatmap-title">做题记录</span>
      <span class="heatmap-count">过去一年共 {{ totalDays }} 天有做题记录</span>
    </div>
    <div class="heatmap-grid" ref="gridRef">
      <div v-for="(week, wi) in weeks" :key="wi" class="heatmap-week">
        <div
          v-for="(day, di) in week"
          :key="di"
          class="heatmap-cell"
          :style="{ background: cellColor(day.count) }"
          :title="day.date ? `${day.date}: ${day.count} 次提交` : ''"
        />
      </div>
    </div>
    <div class="heatmap-legend">
      <span class="legend-label">少</span>
      <div v-for="i in 5" :key="i" class="heatmap-cell legend-cell" :style="{ background: cellColor(i - 1) }" />
      <span class="legend-label">多</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/index'

const data = ref([])

onMounted(async () => {
  try {
    const res = await http.get('/user/heatmap')
    data.value = res.data?.items || []
  } catch {
    data.value = []
  }
})

const countMap = computed(() => {
  const map = {}
  data.value.forEach(item => {
    map[item.date] = item.count
  })
  return map
})

const totalDays = computed(() => data.value.filter(d => d.count > 0).length)

const weeks = computed(() => {
  const today = new Date()
  const start = new Date(today)
  start.setDate(start.getDate() - 364)
  // Align to Sunday
  const dayOfWeek = start.getDay()
  start.setDate(start.getDate() - dayOfWeek)

  const result = []
  let current = new Date(start)

  while (current <= today) {
    const week = []
    for (let i = 0; i < 7; i++) {
      const dateStr = formatDate(current)
      const inRange = current <= today
      week.push({
        date: inRange ? dateStr : '',
        count: countMap.value[dateStr] || 0
      })
      current.setDate(current.getDate() + 1)
    }
    result.push(week)
  }
  return result
})

function formatDate(d) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function cellColor(count) {
  if (!count || count <= 0) return 'var(--heatmap-empty)'
  if (count <= 1) return 'var(--heatmap-L1)'
  if (count <= 3) return 'var(--heatmap-L2)'
  if (count <= 5) return 'var(--heatmap-L3)'
  return 'var(--heatmap-L4)'
}
</script>

<style scoped>
.heatmap-container {
  padding: 4px 0;
}

.heatmap-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.heatmap-title {
  font-size: 17px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.heatmap-count {
  font-size: 12px;
  color: var(--text-muted);
}

.heatmap-grid {
  display: flex;
  gap: 3px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.heatmap-week {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.heatmap-cell {
  width: 13px;
  height: 13px;
  border-radius: var(--radius-xs);
  background: var(--heatmap-empty);
  transition: transform var(--transition-fast);
}

.heatmap-cell:hover {
  transform: scale(1.3);
}

.heatmap-legend {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 10px;
  justify-content: flex-end;
}

.legend-label {
  font-size: 11px;
  color: var(--text-muted);
}

.legend-cell {
  width: 12px;
  height: 12px;
}
</style>
