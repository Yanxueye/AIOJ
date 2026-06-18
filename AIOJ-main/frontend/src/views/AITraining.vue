<template>
  <div class="ai-training-page">
    <div class="training-sidebar">
      <div class="sidebar-header">
        <el-button type="primary" size="small" @click="handleNewChat">
          <el-icon><Plus /></el-icon>新对话
        </el-button>
      </div>
      <el-divider />
      <div class="history-list" v-loading="historyLoading">
        <div
          v-for="conv in aiStore.conversations"
          :key="conv.id"
          class="history-item"
          :class="{ active: conv.id === aiStore.currentConversationId }"
          @click="switchConversation(conv.id)"
        >
          <div class="history-item-main">
            <div class="history-title">{{ conv.title || '新对话' }}</div>
            <div class="history-meta">
              <span class="history-time">{{ formatTime(conv.createdAt) }}</span>
              <span class="history-count">{{ conv.messageCount || 0 }} 条消息</span>
            </div>
          </div>
          <el-button
            class="history-delete-btn"
            text
            size="small"
            :icon="Delete"
            type="danger"
            @click.stop="handleDelete(conv.id)"
          />
        </div>
        <el-empty v-if="!aiStore.conversations.length" description="暂无记录" :image-size="40" />
      </div>
    </div>

    <div class="training-main">
      <AIChat :problem-context="null" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAIStore } from '@/stores/ai'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import AIChat from '@/components/AIChat.vue'

const aiStore = useAIStore()
const historyLoading = ref(false)

async function handleNewChat() {
  aiStore.startNewConversation()
  historyLoading.value = true
  try { await aiStore.loadHistory() } catch {} finally { historyLoading.value = false }
}

async function switchConversation(id) {
  historyLoading.value = true
  try { await aiStore.loadMessages(id) } catch {}
  finally { historyLoading.value = false }
}

async function handleDelete(id) {
  try {
    await ElMessageBox.confirm('确定删除该对话？', '删除确认', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
    await aiStore.deleteConversation(id)
    ElMessage.success('已删除')
  } catch { /* cancelled */ }
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) + ' ' +
    d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  historyLoading.value = true
  try { await aiStore.loadHistory() } catch {}
  finally { historyLoading.value = false }
  if (aiStore.currentMessages.length === 0) {
    aiStore.startNewConversation()
  }
})
</script>

<style scoped>
.ai-training-page {
  display: flex;
  height: calc(100vh - 60px);
}
.training-sidebar {
  width: 300px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  padding: 20px;
  overflow-y: auto;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
  flex-shrink: 0;
}
.sidebar-header h3 {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.history-list {
  flex: 1;
  overflow-y: auto;
}
.history-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  gap: 6px;
}
.history-item:hover { background: var(--bg-warm, rgba(0,0,0,0.03)) }
.history-item.active { background: var(--accent-primary-bg, rgba(99,102,241,0.08)) }
.history-item-main {
  flex: 1;
  min-width: 0;
}
.history-title {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 2px;
}
.history-meta {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--text-muted);
}
.history-delete-btn {
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s;
}
.history-item:hover .history-delete-btn {
  opacity: 1;
}

.training-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.training-main :deep(.ai-chat) {
  border: none;
  border-radius: 0;
  height: 100%;
}
</style>
