<template>
  <div class="ai-chat">
    <div class="chat-header">
      <div class="chat-title">
        <el-icon><MagicStick /></el-icon>
        <span>AI 助手</span>
      </div>
      <div class="chat-actions">
        <el-button text size="small" @click="toggleHistory">
          <el-icon><Clock /></el-icon>
        </el-button>
        <el-button text size="small" @click="handleClear">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- History Panel -->
    <div v-if="showHistory" class="chat-history-panel">
      <div class="history-header">
        <span>历史会话</span>
        <el-button text size="small" @click="showHistory = false"><el-icon><Close /></el-icon></el-button>
      </div>
      <div v-if="aiStore.conversations.length === 0" class="history-empty">暂无历史会话</div>
      <div v-else class="history-list">
        <div
          v-for="conv in aiStore.conversations"
          :key="conv.id"
          class="history-item"
          :class="{ active: conv.id === aiStore.currentConversationId }"
          @click="loadConversation(conv.id)"
        >
          <div class="history-title">{{ conv.title }}</div>
          <div class="history-meta">{{ conv.messageCount }} 条消息</div>
        </div>
      </div>
    </div>

    <div ref="messagesRef" class="chat-messages">
      <div v-if="messages.length === 0" class="chat-empty">
        <el-icon :size="48" :style="{ color: 'var(--text-muted)' }"><ChatDotRound /></el-icon>
        <p>有什么算法问题？向 AI 助手提问吧！</p>
      </div>
      <div
        v-for="msg in visibleMessages"
        :key="msg.id"
        :class="['message', `message-${msg.role}`]"
      >
        <div class="message-avatar">
          <el-avatar v-if="msg.role === 'user'" :size="28" style="background: var(--accent-blue)">我</el-avatar>
          <el-avatar v-else-if="msg.role === 'assistant'" :size="28" style="background: var(--accent-purple)">AI</el-avatar>
        </div>
        <div class="message-content">
          <MarkdownRenderer v-if="msg.role === 'assistant'" :content="msg.content" />
          <div v-else class="user-text">{{ msg.content }}</div>
        </div>
      </div>
      <div v-if="aiStore.loading" class="message message-assistant">
        <div class="message-avatar">
          <el-avatar :size="28" style="background: var(--accent-purple)">AI</el-avatar>
        </div>
        <div class="message-content">
          <div class="typing-indicator">
            <span></span><span></span><span></span>
          </div>
        </div>
      </div>
    </div>

    <!-- Attached problem tags -->
    <div v-if="attachedProblems.length" class="attached-row">
      <div
        v-for="p in attachedProblems"
        :key="p.id"
        class="attached-tag"
      >
        <span class="attached-tag-id">#{{ p.id }}</span>
        <span class="attached-tag-title">{{ p.title }}</span>
        <el-icon class="attached-tag-close" :size="14" @click="removeProblem(p.id)"><Close /></el-icon>
      </div>
    </div>

    <!-- Input area -->
    <div class="chat-input-area">
      <div class="chat-input">
        <el-popover
          v-model:visible="showProblemPop"
          placement="top-start"
          :width="280"
          trigger="manual"
          :hide-after="0"
        >
          <template #reference>
            <el-button
              class="attach-btn"
              :icon="CirclePlus"
              circle
              :disabled="attachedProblems.length >= 3"
              @click="showProblemPop = true"
            />
          </template>
          <div class="pop-input">
            <el-input
              v-model="problemInput"
              placeholder="输入题号，如 1001"
              size="small"
              @keyup.enter="addProblem"
            >
              <template #append>
                <el-button :loading="problemLoading" @click="addProblem">关联</el-button>
              </template>
            </el-input>
            <div v-if="attachedProblems.length >= 3" class="pop-hint">最多关联 3 道题目</div>
          </div>
        </el-popover>
        <el-input
          v-model="inputText"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 4 }"
          placeholder="输入你的问题... (Enter 发送, Shift+Enter 换行)"
          resize="none"
          @keydown.enter.exact.prevent="handleSend"
        />
        <el-button
          type="primary"
          :icon="Promotion"
          circle
          :disabled="!inputText.trim() || aiStore.loading"
          @click="handleSend"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { Promotion, CirclePlus, Close } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAIStore } from '@/stores/ai'
import http from '@/api/index'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps({
  problemContext: { type: Object, default: null },
  codeContext: { type: Object, default: null }
})

const aiStore = useAIStore()
const inputText = ref('')
const messagesRef = ref(null)
const showHistory = ref(false)

// Problem attachment
const attachedProblems = ref([])  // [{id, title, tags}]
const showProblemPop = ref(false)
const problemInput = ref('')
const problemLoading = ref(false)

const messages = computed(() => aiStore.currentMessages)
const visibleMessages = computed(() => messages.value.filter(m => m.role !== 'system'))

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

watch(() => messages.value.length, scrollToBottom)

async function addProblem() {
  const id = parseInt(problemInput.value.trim(), 10)
  if (!id) return
  if (attachedProblems.value.length >= 3) {
    ElMessage.warning('最多关联 3 道题目')
    return
  }
  if (attachedProblems.value.find(p => p.id === id)) {
    ElMessage.warning('已关联此题')
    return
  }
  problemLoading.value = true
  try {
    const r = await http.get(`/problems/${id}`)
    const p = r.data
    if (p?.id) {
      attachedProblems.value.push({
        id: p.id,
        title: p.title,
        tags: p.tags || [],
      })
      problemInput.value = ''
      showProblemPop.value = false
    } else {
      ElMessage.error('题目不存在')
    }
  } catch {
    ElMessage.error('题目不存在')
  } finally {
    problemLoading.value = false
  }
}

function removeProblem(id) {
  attachedProblems.value = attachedProblems.value.filter(p => p.id !== id)
}

async function handleSend() {
  const text = inputText.value.trim()
  if (!text || aiStore.loading) return
  inputText.value = ''
  scrollToBottom()
  const ctx = {
    problem: props.problemContext,
    attachedProblems: attachedProblems.value.map(p => p.id),
    code: props.codeContext,
  }
  await aiStore.sendMessage(text, ctx)
  scrollToBottom()
}

function handleClear() {
  aiStore.clearMessages()
  attachedProblems.value = []
}

async function toggleHistory() {
  showHistory.value = !showHistory.value
  if (showHistory.value && aiStore.conversations.length === 0) {
    await aiStore.loadHistory()
  }
}

async function loadConversation(convId) {
  await aiStore.loadMessages(convId)
  showHistory.value = false
  scrollToBottom()
}
</script>

<style scoped>
.ai-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-hover);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  position: relative;
}
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-card);
  flex-shrink: 0;
}
.chat-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
}
.chat-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.chat-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted);
}
.chat-empty p { font-size: 14px; }
.message {
  display: flex;
  gap: 10px;
  max-width: 100%;
}
.message-user {
  flex-direction: row-reverse;
}
.message-user .message-content {
  background: var(--gradient-amber);
  color: #fff;
  border-radius: 12px 2px 12px 12px;
}
.message-assistant .message-content {
  background: var(--bg-card);
  border: 1px solid var(--border-light);
  border-radius: 2px 12px 12px 12px;
}
.message-content {
  padding: 10px 14px;
  max-width: 85%;
  word-break: break-word;
}
.user-text {
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
}

/* Attached problem tags */
.attached-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 16px 0;
  border-top: 1px solid var(--border-color);
}
.attached-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: var(--accent-primary-bg, rgba(99,102,241,0.1));
  border: 1px solid var(--accent-primary, rgba(99,102,241,0.3));
  border-radius: 999px;
  font-size: 12px;
  cursor: default;
}
.attached-tag-id {
  font-weight: 700;
  color: var(--accent-primary);
  font-family: var(--font-mono);
}
.attached-tag-title {
  color: var(--text-secondary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.attached-tag-close {
  cursor: pointer;
  color: var(--text-muted);
  flex-shrink: 0;
}
.attached-tag-close:hover { color: var(--accent-red); }

/* Input area */
.chat-input-area {
  border-top: 1px solid var(--border-color);
  background: var(--bg-card);
  flex-shrink: 0;
}
.chat-input {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 16px 12px;
}
.chat-input :deep(.el-textarea__inner) {
  box-shadow: none;
  border-radius: var(--radius-sm);
}
.attach-btn {
  flex-shrink: 0;
}
.pop-input {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.pop-hint {
  font-size: 11px;
  color: var(--text-muted);
}

.typing-indicator {
  display: flex; gap: 4px; padding: 4px 0;
}
.typing-indicator span {
  width: 8px; height: 8px;
  background: var(--accent-gold);
  border-radius: 50%;
  animation: typing 1.4s ease-in-out infinite;
}
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }
@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-6px); opacity: 1; }
}

/* History Panel */
.chat-history-panel {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--bg-card);
  z-index: 10;
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-sm);
}
.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  font-weight: 600;
  font-size: 14px;
}
.history-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 13px;
}
.history-list { flex: 1; overflow-y: auto; padding: 8px; }
.history-item {
  padding: 10px 12px; border-radius: 6px; cursor: pointer;
  transition: background 0.15s;
}
.history-item:hover { background: var(--bg-hover); }
.history-item.active { background: var(--accent-primary-bg); }
.history-title {
  font-size: 13px; font-weight: 600; color: var(--text-primary);
  margin-bottom: 2px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.history-meta { font-size: 11px; color: var(--text-muted); }
</style>
