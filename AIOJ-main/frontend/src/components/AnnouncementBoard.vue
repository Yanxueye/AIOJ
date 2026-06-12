<template>
  <div class="announcement-board card">
    <div class="card-title">
      <el-icon><Bell /></el-icon>
      公告栏
    </div>
    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else class="announcement-list">
      <div
        v-for="item in announcements"
        :key="item.id"
        class="announcement-item"
        @click="showDetail(item)"
      >
        <el-tag :type="item.type" size="small" effect="dark">
          {{ item.type === 'success' ? '公告' : item.type === 'warning' ? '通知' : item.type === 'info' ? '更新' : '活动' }}
        </el-tag>
        <span class="announcement-title">{{ item.title }}</span>
        <span class="announcement-date">{{ item.date }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessageBox } from 'element-plus'
import { problemApi } from '@/api/problem'

const announcements = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await problemApi.getAnnouncements()
    announcements.value = res.data
  } finally {
    loading.value = false
  }
})

function showDetail(item) {
  ElMessageBox.alert(item.content, item.title, {
    confirmButtonText: '知道了',
    dangerouslyUseHTMLString: false
  })
}
</script>

<style scoped>
.announcement-board .card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.announcement-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.announcement-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 14px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.announcement-item:hover {
  background: var(--bg-hover);
}

.announcement-title {
  flex: 1;
  font-size: 13.5px;
  font-weight: 500;
}

.announcement-date {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}

.loading-wrap {
  padding: 16px 0;
}
</style>
