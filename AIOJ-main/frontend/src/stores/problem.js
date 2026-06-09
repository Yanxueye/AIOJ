import { defineStore } from 'pinia'
import { ref } from 'vue'
import { problemApi } from '@/api/problem'

export const useProblemStore = defineStore('problem', () => {
  const problems = ref([])
  const currentProblem = ref(null)
  const total = ref(0)
  const loading = ref(false)

  async function fetchProblems(params = {}) {
    loading.value = true
    try {
      const res = await problemApi.getList(params)
      problems.value = res.data.list
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  async function fetchProblem(id) {
    loading.value = true
    try {
      const res = await problemApi.getDetail(id)
      currentProblem.value = res.data
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchAdminProblem(id) {
    loading.value = true
    try {
      const res = await problemApi.getAdminDetail(id)
      currentProblem.value = res.data
      return res.data
    } finally {
      loading.value = false
    }
  }

  return { problems, currentProblem, total, loading, fetchProblems, fetchProblem, fetchAdminProblem }
})
