import { defineStore } from 'pinia'
import { ref } from 'vue'
import { aiApi } from '@/api/ai'

export const useAIStore = defineStore('ai', () => {
  const conversations = ref([])
  const currentMessages = ref([])
  const loading = ref(false)
  const currentConversationId = ref(null)

  function addMessage(role, content) {
    currentMessages.value.push({
      id: `${Date.now()}_${Math.random().toString(36).slice(2)}`,
      role,
      content,
      timestamp: new Date().toISOString()
    })
  }

  async function sendMessage(content, problemContext = null) {
    addMessage('user', content)
    loading.value = true
    try {
      const res = await aiApi.chat({
        message: content,
        history: currentMessages.value.slice(0, -1).map(m => ({
          role: m.role,
          content: m.content
        })),
        problem_id: problemContext?.id || null,
        conversation_id: currentConversationId.value || ''
      })
      currentConversationId.value = res.data.conversationId || currentConversationId.value
      addMessage('assistant', res.data.reply)
      return res.data.reply
    } catch (err) {
      addMessage('assistant', '抱歉，AI 服务暂时不可用，请稍后重试。')
      throw err
    } finally {
      loading.value = false
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

  async function diagnoseCode({ problemId, language, code, submissionId = 0, judgeStatus = '', errorMessage = '' }) {
    addMessage('user', '请诊断当前代码并指出潜在错误。')
    loading.value = true
    try {
      const res = await aiApi.diagnoseCode({ problemId, language, code, submissionId, judgeStatus, errorMessage })
      addMessage('assistant', res.data.rawMarkdown || formatDiagnosis(res.data))
      return res.data
    } catch (err) {
      addMessage('assistant', '代码诊断服务暂时不可用，请稍后重试。')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function solveProblem({ problemId, question = '', level = 'hint' }) {
    addMessage('user', level === 'hint' ? '请给我这道题的解题提示。' : '请讲解这道题的解法。')
    loading.value = true
    try {
      const res = await aiApi.solveProblem({ problemId, question, level })
      addMessage('assistant', formatSolve(res.data))
      return res.data
    } catch (err) {
      addMessage('assistant', '解题服务暂时不可用，请稍后重试。')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function buildKnowledgeGraph({ problemId = null, scope = 'recent' } = {}) {
    addMessage('user', problemId ? '请基于当前题目整理我的知识图谱。' : '请基于最近做题记录整理我的知识图谱。')
    loading.value = true
    try {
      const res = await aiApi.buildKnowledgeGraph({ problemId, scope })
      addMessage('assistant', res.data.rawMarkdown || formatKnowledgeGraph(res.data))
      return res.data
    } catch (err) {
      addMessage('assistant', '知识图谱服务暂时不可用，请稍后重试。')
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearMessages() {
    currentMessages.value = []
    currentConversationId.value = null
  }

  function startNewConversation(problemContext = null) {
    clearMessages()
    if (problemContext) {
      addMessage('system', `当前题目上下文：[${problemContext.id}] ${problemContext.title}`)
    }
  }

  function formatDiagnosis(data = {}) {
    const issues = (data.issues || []).map(i => `- **${i.severity || 'info'}**：${i.message || ''}${i.hint ? `。${i.hint}` : ''}`).join('\n')
    const suggestions = (data.suggestions || []).map(s => `- ${s}`).join('\n')
    return `### 代码诊断\n\n${data.summary || ''}\n\n#### 发现的问题\n\n${issues || '- 暂未发现明显问题'}\n\n#### 建议\n\n${suggestions || '- 补充边界用例继续验证'}`
  }

  function formatSolve(data = {}) {
    const hints = (data.hints || []).map(h => `- ${h}`).join('\n')
    return `${data.answer || ''}${hints ? `\n\n#### 提示\n\n${hints}` : ''}${data.complexity ? `\n\n#### 复杂度\n\n${data.complexity}` : ''}`
  }

  function formatKnowledgeGraph(data = {}) {
    return `### 学习知识图谱\n\n${data.summary || ''}\n\n- 节点数：${data.nodes?.length || 0}\n- 关系数：${data.edges?.length || 0}`
  }

  return {
    conversations, currentMessages, loading, currentConversationId,
    sendMessage, loadHistory, loadMessages, diagnoseCode, solveProblem, buildKnowledgeGraph,
    clearMessages, startNewConversation, addMessage
  }
})
