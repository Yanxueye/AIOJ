<template>
  <div class="ai-chat">
    <div class="chat-header">
      <div class="chat-title">
        <el-icon><MagicStick /></el-icon>
        <span>AI 助手</span>
        <el-tag v-if="problemContext" size="small" type="info" closable @close="$emit('clear-context')">
          #{{ problemContext.id }}
        </el-tag>
      </div>
      <div class="chat-actions">
        <el-button v-if="problemContext" text size="small" :disabled="aiStore.loading" @click="handleHint">
          解题提示
        </el-button>
        <el-button v-if="problemContext" text size="small" :disabled="aiStore.loading" @click="handleDiagnose">
          诊断代码
        </el-button>
        <el-button text size="small" @click="handleClear">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
    </div>

    <div ref="messagesRef" class="chat-messages">
      <div v-if="messages.length === 0" class="chat-empty">
        <el-icon :size="48" color="#c0c4cc"><ChatDotRound /></el-icon>
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

    <div class="chat-input">
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
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { Promotion } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAIStore } from '@/stores/ai'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps({
  problemContext: { type: Object, default: null },
  codeContext: { type: Object, default: null }
})

defineEmits(['clear-context'])

const aiStore = useAIStore()
const inputText = ref('')
const messagesRef = ref(null)

const messages = computed(() => aiStore.currentMessages)
const visibleMessages = computed(() => messages.value.filter(m => m.role !== 'system'))
const canDiagnose = computed(() => Boolean(props.problemContext?.id && props.codeContext?.code?.trim()))

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

watch(() => messages.value.length, scrollToBottom)

async function handleSend() {
  const text = inputText.value.trim()
  if (!text || aiStore.loading) return
  inputText.value = ''
  scrollToBottom()
  await aiStore.sendMessage(text, props.problemContext)
  scrollToBottom()
}

function handleClear() {
  aiStore.clearMessages()
}

async function handleHint() {
  if (!props.problemContext?.id || aiStore.loading) return
  try {
    await aiStore.solveProblem({ problemId: props.problemContext.id, level: 'hint' })
  } catch {
    return
  }
  scrollToBottom()
}

async function handleDiagnose() {
  if (!canDiagnose.value) {
    ElMessage.warning('请先输入代码')
    return
  }
  try {
    await aiStore.diagnoseCode({
      problemId: props.problemContext.id,
      language: props.codeContext.language || 'cpp',
      code: props.codeContext.code
    })
  } catch {
    return
  }
  scrollToBottom()
}
</script>

<style scoped>
.ai-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fafbfc;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  background: #fff;
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
.chat-empty p {
  font-size: 14px;
}
.message {
  display: flex;
  gap: 10px;
  max-width: 100%;
}
.message-user {
  flex-direction: row-reverse;
}
.message-user .message-content {
  background: var(--accent-blue);
  color: #fff;
  border-radius: 12px 2px 12px 12px;
}
.message-assistant .message-content {
  background: #fff;
  border: 1px solid var(--border-color);
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
.chat-input {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  background: #fff;
}
.chat-input :deep(.el-textarea__inner) {
  box-shadow: none;
  border-radius: var(--radius-sm);
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 4px 0;
}
.typing-indicator span {
  width: 8px;
  height: 8px;
  background: var(--accent-purple);
  border-radius: 50%;
  animation: typing 1.4s ease-in-out infinite;
}
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }
@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-6px); opacity: 1; }
}
</style>
