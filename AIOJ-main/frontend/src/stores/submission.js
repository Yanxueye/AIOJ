import { defineStore } from 'pinia'
import { ref } from 'vue'
import { submissionApi } from '@/api/submission'

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
    return res.data
  }

  async function getDetail(id) {
    const res = await submissionApi.getDetail(id)
    return res.data
  }

  return { submissions, total, loading, currentResult, fetchSubmissions, submit, getDetail }
})
