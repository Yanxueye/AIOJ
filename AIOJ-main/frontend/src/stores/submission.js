import { defineStore } from 'pinia'
import { ref } from 'vue'
import { submissionApi } from '@/api/submission'

const TERMINAL_STATUSES = new Set([
  'Accepted',
  'Wrong Answer',
  'Time Limit Exceeded',
  'Runtime Error',
  'Compile Error',
  'Memory Limit Exceeded',
  'Output Limit Exceeded',
  'System Error'
])

export const useSubmissionStore = defineStore('submission', () => {
  const submissions = ref([])
  const total = ref(0)
  const loading = ref(false)
  const currentResult = ref(null)

  async function fetchSubmissions(params = {}) {
    loading.value = true
    try {
      const res = await submissionApi.getList(params)
      submissions.value = res.data.list
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  async function submit(data) {
    const res = await submissionApi.submit(data)
    currentResult.value = res.data
    if (!res.data?.id || TERMINAL_STATUSES.has(res.data.status)) {
      return res.data
    }
    const finalResult = await waitForResult(res.data.id, res.data.traceId)
    currentResult.value = finalResult
    return finalResult
  }

  async function getDetail(id) {
    const res = await submissionApi.getDetail(id)
    return res.data
  }

  async function waitForResult(id, traceId = '', maxAttempts = 20, intervalMs = 800) {
    let latest = null
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise(resolve => setTimeout(resolve, i === 0 ? 300 : intervalMs))
      try {
        latest = await getDetail(id)
      } catch (error) {
        if (error?.response?.status === 404) {
          continue
        }
        throw error
      }
      if (traceId && latest?.traceId && latest.traceId !== traceId) {
        continue
      }
      if (TERMINAL_STATUSES.has(latest?.status)) {
        return latest
      }
    }
    return latest || currentResult.value
  }

  return { submissions, total, loading, currentResult, fetchSubmissions, submit, getDetail }
})
