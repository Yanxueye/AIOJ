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

  async function runProblem(id, data) {
    const res = await problemApi.runCode(id, data)
    return res.data
  }

  async function favoriteProblem(id) {
    const res = await problemApi.favorite(id)
    if (currentProblem.value?.id === Number(id)) {
      currentProblem.value = { ...currentProblem.value, favorite: true }
    }
    problems.value = problems.value.map(item => item.id === Number(id) ? { ...item, favorite: true } : item)
    return res.data
  }

  async function unfavoriteProblem(id) {
    const res = await problemApi.unfavorite(id)
    if (currentProblem.value?.id === Number(id)) {
      currentProblem.value = { ...currentProblem.value, favorite: false }
    }
    problems.value = problems.value.map(item => item.id === Number(id) ? { ...item, favorite: false } : item)
    return res.data
  }

  return { problems, currentProblem, total, loading, fetchProblems, fetchProblem, fetchAdminProblem, runProblem, favoriteProblem, unfavoriteProblem }
})
