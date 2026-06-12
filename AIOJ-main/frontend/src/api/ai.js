import http from './index'

// AI endpoints need longer timeouts (LLM calls take 1-2 minutes)
const aiHttp = {
  post: (url, data) => http.post(url, data, { timeout: 180000 }),
  get: (url) => http.get(url, { timeout: 30000 })
}

export const aiApi = {
  chat: data => aiHttp.post('/ai/chat', data),
  getHistory: () => aiHttp.get('/ai/history'),
  getMessages: id => aiHttp.get(`/ai/conversations/${id}/messages`),
  diagnoseCode: data => aiHttp.post('/ai/code-diagnosis', data),
  buildKnowledgeGraph: data => aiHttp.post('/ai/knowledge-graph', data),
  solveProblem: data => aiHttp.post('/ai/solve', data)
}
