<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { toast } from 'vue-sonner'
import { useProjectStore } from '@/stores/project'
import { useRunStore } from '@/stores/run'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { listConfigs } from '@/api/configs'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import SummaryCards from '@/components/SummaryCards.vue'
import type { SummaryCard } from '@/components/SummaryCards.vue'
import type { EvalRun } from '@/types'

const projectStore = useProjectStore()
const runStore = useRunStore()
const { currentProjectName } = storeToRefs(projectStore)
const { recentRuns, activeRuns } = storeToRefs(runStore)
const router = useRouter()
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

async function loadData() {
  if (!currentProjectName.value) return
  try {
    await runStore.refresh(currentProjectName.value)
    const configs = await listConfigs(currentProjectName.value)
    configCount.value = configs.length
  } catch {
    toast.error('Failed to load dashboard data')
  }
}

watch(currentProjectName, loadData, { immediate: true })

function handleRowClick(row: EvalRun) {
  router.push({ name: 'result-detail', params: { id: row.id } })
}

function goToNewRun() {
  router.push({ name: 'runs' })
}

function formatPassRate(run: EvalRun): string {
  return `${((run.summary?.overall_pass_rate ?? 0) * 100).toFixed(1)}%`
}

function formatDuration(run: EvalRun): string {
  return `${((run.duration_ms ?? 0) / 1000).toFixed(1)}s`
}

function formatDate(run: EvalRun): string {
  return run.started_at ? new Date(run.started_at).toLocaleString() : '-'
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">Dashboard</h1>
        <p class="text-sm text-muted-foreground mt-1">Overview of your evaluations</p>
      </div>
      <Button :disabled="!currentProjectName" @click="goToNewRun">New Run</Button>
    </div>

    <template v-if="currentProjectName">
      <SummaryCards :cards="summaryCards" />

      <div>
        <h3 class="text-base font-medium text-zinc-900 mb-3">Recent Runs</h3>
        <div v-if="recentRuns.length > 0" class="bg-white rounded-lg border border-gray-200 shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Suite</TableHead>
                <TableHead>Agent</TableHead>
                <TableHead>Pass Rate</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="run in recentRuns.slice(0, 10)"
                :key="run.id"
                class="cursor-pointer hover:bg-zinc-50"
                @click="handleRowClick(run)"
              >
                <TableCell class="font-medium">{{ run.suite_name }}</TableCell>
                <TableCell>{{ run.agent_type }}</TableCell>
                <TableCell>{{ formatPassRate(run) }}</TableCell>
                <TableCell>{{ formatDuration(run) }}</TableCell>
                <TableCell class="text-muted-foreground">{{ formatDate(run) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <div v-else class="flex items-center justify-center py-12 text-sm text-muted-foreground">
          No runs yet. Start one from the Runs page.
        </div>
      </div>
    </template>

    <div v-else class="flex items-center justify-center py-12 text-sm text-muted-foreground">
      No project selected. Add a project to get started.
    </div>
  </div>
</template>
