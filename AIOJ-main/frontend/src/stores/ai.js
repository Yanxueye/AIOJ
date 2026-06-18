import { defineStore } from 'pinia'
import { ref } from 'vue'
import { aiApi } from '@/api/ai'

export const useAIStore = defineStore('ai', () => {
  const conversations = ref([])
  const currentMessages = ref([])
  const loading = ref(false)
  const chatLoading = ref(false)
  const diagnoseLoading = ref(false)
  const solveLoading = ref(false)
  const error = ref(null)
  const currentConversationId = ref(null)

  function addMessage(role, content) {
    currentMessages.value.push({
      id: `${Date.now()}_${Math.random().toString(36).slice(2)}`,
      role,
      content,
      timestamp: new Date().toISOString()
    })
  }

  // Unified sendMessage — the only send method needed.
  // context: { problem, attachedProblems: [id, ...], code: { language, code } }
  async function sendMessage(content, context = null, mode = 'chat') {
    const ctx = context || {}
    addMessage('user', content)
    loading.value = true
    chatLoading.value = true
    error.value = null
    try {
      const res = await aiApi.chat({
        mode,
        message: content,
        history: currentMessages.value.slice(0, -1).map(m => ({
          role: m.role,
          content: m.content
        })),
        problem_id: ctx.problem?.id || null,
        problem_ids: ctx.attachedProblems || [],
        conversation_id: currentConversationId.value || '',
        code_language: ctx.code?.language || null,
        code: ctx.code?.code || null
      })
      currentConversationId.value = res.data.conversationId || currentConversationId.value
      addMessage('assistant', res.data.reply)
      return res.data.reply
    } catch (err) {
      error.value = 'AI 服务暂时不可用，请稍后重试。'
      addMessage('assistant', '抱歉，AI 服务暂时不可用，请稍后重试。')
      throw err
    } finally {
      loading.value = false
      chatLoading.value = false
    }
  }

  async function loadHistory() {
    const res = await aiApi.getHistory()
    conversations.value = res.data.conversations || []
  }

  async function loadMessages(conversationId) {
    const res = await aiApi.getMessages(conversationId)
    currentConversationId.value = conversationId
    currentMessages.value = (res.data.messages || []).map(m => ({
      id: m.id,
      role: m.role,
      content: m.content,
      timestamp: m.createdAt
    }))
    return currentMessages.value
  }

  // Specialized methods — all delegate to sendMessage with appropriate mode and context object
  async function diagnoseCode({ problemId, language, code, submissionId = 0, judgeStatus = '', errorMessage = '' }) {
    return sendMessage('请诊断当前代码并指出潜在错误。', { problem: { id: problemId }, code: { language, code } }, 'code-diagnosis')
  }

  async function solveProblem({ problemId, question = '', level = 'hint', language = '', code = '' }) {
    const msg = level === 'hint' ? '请给我这道题的解题提示。' : '请讲解这道题的解法。'
    return sendMessage(msg, { problem: { id: problemId }, code: { language, code } }, 'solve')
  }

  async function buildKnowledgeGraph({ problemId = null, scope = 'recent' } = {}) {
    const msg = problemId ? '请基于当前题目整理我的知识图谱。' : '请基于最近做题记录整理我的知识图谱。'
    return sendMessage(msg, { problem: problemId ? { id: problemId } : null }, 'knowledge-graph')
  }

  function clearMessages() {
    currentMessages.value = []
    currentConversationId.value = null
    error.value = null
  }

  function startNewConversation(problemContext = null) {
    clearMessages()
    if (problemContext) {
      addMessage('system', `当前题目上下文：[${problemContext.id}] ${problemContext.title}`)
    }
  }

  async function deleteConversation(id) {
    await aiApi.deleteConversation(id)
    conversations.value = conversations.value.filter(c => c.id !== id)
    if (currentConversationId.value === id) {
      clearMessages()
    }
  }

  return {
    conversations, currentMessages, loading, chatLoading, diagnoseLoading, solveLoading, error, currentConversationId,
    sendMessage, loadHistory, loadMessages, diagnoseCode, solveProblem, buildKnowledgeGraph,
    clearMessages, startNewConversation, deleteConversation, addMessage
  }
})
