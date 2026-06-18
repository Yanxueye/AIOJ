import http from './index'

// AI endpoints need longer timeouts (LLM calls take 1-2 minutes)
const aiHttp = {
  post: (url, data) => http.post(url, data, { timeout: 180000 }),
  get: (url) => http.get(url, { timeout: 30000 })
}

export const aiApi = {
  // Unified chat — the only AI endpoint needed (use mode field for different scenarios)
  chat: data => aiHttp.post('/ai/chat', data),
  // Conversation management
  getHistory: () => aiHttp.get('/ai/history'),
  getMessages: id => aiHttp.get(`/ai/conversations/${id}/messages`),
  deleteConversation: id => http.delete(`/ai/conversations/${id}`, { timeout: 10000 })
}
