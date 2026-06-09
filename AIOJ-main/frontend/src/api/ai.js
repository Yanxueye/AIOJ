import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const aiApi = {
  chat: data => USE_MOCK ? mockApi.aiChat(data) : http.post('/ai/chat', data),
  getHistory: () => USE_MOCK ? mockApi.getAIHistory() : http.get('/ai/history'),
  getMessages: id => USE_MOCK ? mockApi.getAIMessages(id) : http.get(`/ai/conversations/${id}/messages`),
  diagnoseCode: data => USE_MOCK ? mockApi.aiCodeDiagnosis(data) : http.post('/ai/code-diagnosis', data),
  buildKnowledgeGraph: data => USE_MOCK ? mockApi.aiKnowledgeGraph(data) : http.post('/ai/knowledge-graph', data),
  solveProblem: data => USE_MOCK ? mockApi.aiSolve(data) : http.post('/ai/solve', data)
}
