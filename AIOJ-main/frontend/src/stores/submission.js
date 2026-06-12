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
      // Already terminal — add to list immediately
      submissions.value = [res.data, ...submissions.value]
      return res.data
    }
    let finalResult = null
    try {
      finalResult = await waitForResultStream(res.data.id, res.data.traceId)
    } catch {
      finalResult = await waitForResult(res.data.id, res.data.traceId)
    }
    currentResult.value = finalResult
    // Add completed submission to the top of the list
    if (finalResult) {
      submissions.value = [finalResult, ...submissions.value.filter(s => s.id !== finalResult.id)]
    }
    return finalResult
  }

  async function getDetail(id) {
    const res = await submissionApi.getDetail(id)
    return res.data
  }

  async function getCases(id) {
    const res = await submissionApi.getCases(id)
    return res.data
  }

  async function getOutput(id) {
    const res = await submissionApi.getOutput(id)
    return res.data
  }

  async function waitForResult(id, traceId = '', maxAttempts = 120, intervalMs = 500) {
    let latest = null
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise(resolve => setTimeout(resolve, i === 0 ? 200 : intervalMs))
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
      // Update currentResult on each poll so UI shows live status
      if (latest) {
        currentResult.value = latest
      }
      if (TERMINAL_STATUSES.has(latest?.status)) {
        return latest
      }
    }
    return latest || currentResult.value
  }

  function waitForResultStream(id, traceId = '') {
    return new Promise((resolve, reject) => {
      if (typeof EventSource === 'undefined') {
        reject(new Error('EventSource unsupported'))
        return
      }

      const token = localStorage.getItem('toj_token')
      const url = submissionApi.stream(id)
      const es = new EventSource(`${url}?token=${encodeURIComponent(token || '')}`)

      es.addEventListener('submission', event => {
        const data = JSON.parse(event.data)
        if (traceId && data?.traceId && data.traceId !== traceId) {
          return
        }
        currentResult.value = data
        if (TERMINAL_STATUSES.has(data?.status)) {
          es.close()
          resolve(data)
        }
      })

      es.addEventListener('error', () => {
        es.close()
        reject(new Error('stream error'))
      })
    })
  }

  return { submissions, total, loading, currentResult, fetchSubmissions, submit, getDetail, getCases, getOutput }
})
