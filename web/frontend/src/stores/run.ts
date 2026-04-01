import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ActiveRun, EvalRun } from '@/types'
import { listActiveRuns, listRuns } from '@/api/runs'

export const useRunStore = defineStore('run', () => {
  const activeRuns = ref<ActiveRun[]>([])
  const recentRuns = ref<EvalRun[]>([])
  const loading = ref(false)

  async function fetchActiveRuns(project: string) {
    activeRuns.value = await listActiveRuns(project)
  }

  async function fetchRecentRuns(project: string) {
    loading.value = true
    try {
      recentRuns.value = await listRuns(project)
    } finally {
      loading.value = false
    }
  }

  async function refresh(project: string) {
    await Promise.all([fetchActiveRuns(project), fetchRecentRuns(project)])
  }

  return {
    activeRuns,
    recentRuns,
    loading,
    fetchActiveRuns,
    fetchRecentRuns,
    refresh,
  }
})
