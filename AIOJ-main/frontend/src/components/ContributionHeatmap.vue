<template>
  <div class="heatmap-container">
    <div class="heatmap-header">
      <span class="heatmap-title">做题记录</span>
      <span class="heatmap-count">{{ yearRange }} · {{ totalDays }} 天有做题记录</span>
    </div>

    <div class="heatmap-body">
      <!-- Day labels (left) -->
      <div class="day-labels">
        <div class="day-label-space" />
        <div class="day-label">一</div>
        <div class="day-label-space" />
        <div class="day-label">三</div>
        <div class="day-label-space" />
        <div class="day-label">五</div>
        <div class="day-label-space" />
      </div>

      <!-- Grid -->
      <div class="grid-area">
        <!-- Month labels -->
        <div class="month-row">
          <div
            v-for="(m, i) in monthLabels"
            :key="i"
            class="month-label"
            :style="{ width: m.weeks * CELL_STRIDE + 'px' }"
          >{{ m.name }}</div>
        </div>

        <!-- Weeks as columns -->
        <div class="weeks-row">
          <div v-for="(week, wi) in weeks" :key="wi" class="week-col">
            <div
              v-for="(day, di) in week"
              :key="di"
              class="cell"
              :class="day.date ? lvl(day.count) : 'cell-invisible'"
              :title="day.date ? `${day.date}：${day.count} 次提交` : ''"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="legend">
      <span class="legend-text">少</span>
      <div class="cell legend-cell lvl-1" />
      <div class="cell legend-cell lvl-2" />
      <div class="cell legend-cell lvl-3" />
      <div class="cell legend-cell lvl-4" />
      <span class="legend-text">多</span>
      <span class="legend-hint">（1 / 2-3 / 4-5 / 6+）</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/index'

const data = ref([])

// Layout constants — must match CSS exactly
const CELL_SIZE = 13   // px
const GAP = 3          // px between cells
const CELL_STRIDE = CELL_SIZE + GAP  // 16px per week column

onMounted(async () => {
  try {
    const res = await http.get('/user/heatmap')
    data.value = res.data?.items || []
  } catch { data.value = [] }
})

const countMap = computed(() => {
  const m = {}
  data.value.forEach(d => { m[d.date] = d.count })
  return m
})

const totalDays = computed(() => data.value.filter(d => d.count > 0).length)

const yearRange = computed(() => {
  if (!weeks.value.length) return ''
  const first = weeks.value[0].find(d => d.date)
  const last = [...weeks.value[weeks.value.length - 1]].reverse().find(d => d.date)
  if (!first || !last) return ''
  const y1 = new Date(first.date).getFullYear()
  const y2 = new Date(last.date).getFullYear()
  return y1 === y2 ? `${y1}年` : `${y1}年-${y2}年`
})

// Build weeks: past 365 days, align to Sunday
const weeks = computed(() => {
  const today = new Date()
  const start = new Date(today)
  start.setDate(start.getDate() - 364)
  start.setDate(start.getDate() - start.getDay()) // align to Sunday

  const result = []
  const cur = new Date(start)
  while (cur <= today) {
    const week = []
    for (let i = 0; i < 7; i++) {
      const ds = fmt(cur)
      week.push({
        date: cur <= today ? ds : '',
        count: countMap.value[ds] || 0
      })
      cur.setDate(cur.getDate() + 1)
    }
    result.push(week)
  }
  return result
})

// Build month labels with correct week spans
const monthLabels = computed(() => {
  const names = ['1月','2月','3月','4月','5月','6月','7月','8月','9月','10月','11月','12月']
  const labels = []
  let lastMonth = -1
  let span = 0

  weeks.value.forEach(week => {
    const firstDay = week.find(d => d.date)
    if (!firstDay) { span++; return }
    const m = new Date(firstDay.date).getMonth()
    if (m !== lastMonth) {
      if (lastMonth >= 0) labels.push({ name: names[lastMonth], weeks: span })
      lastMonth = m
      span = 1
    } else {
      span++
    }
  })
  if (lastMonth >= 0) labels.push({ name: names[lastMonth], weeks: span })
  return labels
})

function fmt(d) {
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
}

function lvl(count) {
  if (!count || count <= 0) return 'lvl-0'
  if (count === 1) return 'lvl-1'
  if (count <= 3) return 'lvl-2'
  if (count <= 5) return 'lvl-3'
  return 'lvl-4'
}
</script>

<style scoped>
.heatmap-container { padding: 4px 0; }

.heatmap-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 12px;
}
.heatmap-title { font-size: 17px; font-weight: 800; color: var(--text-primary); }
.heatmap-count { font-size: 12px; color: var(--text-muted); }

.heatmap-body { display: flex; gap: 4px; }

/* Day labels */
.day-labels {
  display: flex; flex-direction: column; gap: 3px;
  padding-top: 18px;
}
.day-label { height: 13px; font-size: 10px; color: var(--text-muted); display: flex; align-items: center; justify-content: flex-end; width: 18px; }
.day-label-space { height: 13px; }

.grid-area { overflow-x: auto; }

/* Month labels — width = weeks * 16px (13px cell + 3px gap) */
.month-row { display: flex; height: 18px; margin-bottom: 2px; }
.month-label {
  font-size: 10px; color: var(--text-muted);
  display: flex; align-items: center;
  flex-shrink: 0; overflow: hidden; white-space: nowrap;
}

/* Weeks */
.weeks-row { display: flex; gap: 3px; }
.week-col { display: flex; flex-direction: column; gap: 3px; flex-shrink: 0; }

/* Cell */
.cell { width: 13px; height: 13px; border-radius: 2px; flex-shrink: 0; }
.cell-invisible { background: transparent; }

/* Colors — GitHub style */
.lvl-0 { background: #ebedf0; }
.lvl-1 { background: #9be9a8; }
.lvl-2 { background: #40c463; }
.lvl-3 { background: #30a14e; }
.lvl-4 { background: #216e39; }

[data-theme="dark"] .lvl-0 { background: #161b22; }
[data-theme="dark"] .lvl-1 { background: #0e4429; }
[data-theme="dark"] .lvl-2 { background: #006d32; }
[data-theme="dark"] .lvl-3 { background: #26a641; }
[data-theme="dark"] .lvl-4 { background: #39d353; }

/* Legend */
.legend { display: flex; align-items: center; gap: 3px; margin-top: 10px; justify-content: flex-end; }
.legend-text { font-size: 11px; color: var(--text-muted); }
.legend-hint { font-size: 10px; color: var(--text-muted); margin-left: 4px; }
.legend-cell { width: 13px; height: 13px; }
</style>
