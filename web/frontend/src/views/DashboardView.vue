<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { NSpace, NText, NButton, NDataTable, NTag, NEmpty, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { useRunStore } from '@/stores/run'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { listConfigs } from '@/api/configs'
import SummaryCards from '@/components/SummaryCards.vue'
import type { SummaryCard } from '@/components/SummaryCards.vue'
import type { EvalRun } from '@/types'

const projectStore = useProjectStore()
const runStore = useRunStore()
const { currentProjectName } = storeToRefs(projectStore)
const { recentRuns, activeRuns } = storeToRefs(runStore)
const router = useRouter()
const message = useMessage()
const configCount = ref(0)

const summaryCards = computed<SummaryCard[]>(() => {
  const runs = recentRuns.value
  const avgPassRate = runs.length > 0
    ? runs.reduce((sum, r) => sum + (r.summary?.overall_pass_rate ?? 0), 0) / runs.length
    : 0
  return [
    { label: 'Total Runs', value: runs.length },
    { label: 'Configs', value: configCount.value },
    { label: 'Avg Pass Rate', value: `${(avgPassRate * 100).toFixed(1)}`, suffix: '%' },
    { label: 'Active Runs', value: activeRuns.value.length },
  ]
})

const columns: DataTableColumns<EvalRun> = [
  { title: 'Suite', key: 'suite_name' },
  { title: 'Agent', key: 'agent_type' },
  {
    title: 'Pass Rate', key: 'pass_rate',
    render(row) { return `${((row.summary?.overall_pass_rate ?? 0) * 100).toFixed(1)}%` },
  },
  {
    title: 'Duration', key: 'duration_ms',
    render(row) { return `${((row.duration_ms ?? 0) / 1000).toFixed(1)}s` },
  },
  {
    title: 'Date', key: 'started_at',
    render(row) { return row.started_at ? new Date(row.started_at).toLocaleString() : '-' },
  },
]

async function loadData() {
  if (!currentProjectName.value) return
  try {
    await runStore.refresh(currentProjectName.value)
    const configs = await listConfigs(currentProjectName.value)
    configCount.value = configs.length
  } catch {
    message.error('Failed to load dashboard data')
  }
}

watch(currentProjectName, loadData, { immediate: true })

function handleRowClick(row: EvalRun) {
  router.push({ name: 'result-detail', params: { id: row.id } })
}

function goToNewRun() {
  router.push({ name: 'runs' })
}
</script>

<template>
  <NSpace vertical :size="24">
    <NSpace justify="space-between" align="center">
      <NText tag="h1" style="margin: 0">Dashboard</NText>
      <NButton type="primary" :disabled="!currentProjectName" @click="goToNewRun">
        New Run
      </NButton>
    </NSpace>

    <template v-if="currentProjectName">
      <SummaryCards :cards="summaryCards" />

      <NText tag="h3">Recent Runs</NText>
      <NDataTable
        v-if="recentRuns.length > 0"
        :columns="columns"
        :data="recentRuns.slice(0, 10)"
        :row-props="(row: EvalRun) => ({ style: 'cursor: pointer', onClick: () => handleRowClick(row) })"
      />
      <NEmpty v-else description="No runs yet. Start one from the Runs page." />
    </template>

    <NEmpty v-else description="No project selected. Add a project to get started." />
  </NSpace>
</template>
