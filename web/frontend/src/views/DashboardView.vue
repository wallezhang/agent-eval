<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { toast } from 'vue-sonner'
import { useProjectStore } from '@/stores/project'
import { useRunStore } from '@/stores/run'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { listConfigs } from '@/api/configs'
import { LayoutDashboard } from 'lucide-vue-next'
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

function formatPassRate(run: EvalRun): string {
  return `${((run.summary?.overall_pass_rate ?? 0) * 100).toFixed(1)}%`
}

function formatDuration(run: EvalRun): string {
  return `${((run.duration_ms ?? 0) / 1000).toFixed(1)}s`
}

function formatDate(run: EvalRun): string {
  return run.started_at ? new Date(run.started_at).toLocaleString() : '-'
}

function passRateColor(run: EvalRun): string {
  const rate = run.summary?.overall_pass_rate ?? 0
  if (rate >= 0.8) return 'text-success font-semibold'
  if (rate >= 0.5) return 'text-warning font-semibold'
  return 'text-error font-semibold'
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-extrabold tracking-tight font-display">
          <span class="text-gradient">Dashboard</span>
        </h1>
        <p class="text-sm text-muted-foreground mt-1">Overview of your evaluations</p>
      </div>
    </div>

    <template v-if="currentProjectName">
      <SummaryCards :cards="summaryCards" />

      <div>
        <h3 class="text-base font-semibold text-zinc-900 mb-3 font-display">Recent Runs</h3>
        <div v-if="recentRuns.length > 0" class="bg-white rounded-lg border border-gray-200 card-shadow overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow class="bg-zinc-50/50">
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Suite</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Agent</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Pass Rate</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Duration</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="(run, index) in recentRuns.slice(0, 10)"
                :key="run.id"
                class="cursor-pointer hover:bg-primary-light/30 transition-colors duration-200 group"
                :class="index % 2 === 1 ? 'bg-zinc-50/30' : ''"
                @click="handleRowClick(run)"
              >
                <TableCell class="font-medium">
                  <div class="flex items-center gap-2">
                    <div class="w-0.5 h-4 bg-primary rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
                    {{ run.suite_name }}
                  </div>
                </TableCell>
                <TableCell>{{ run.agent_type }}</TableCell>
                <TableCell :class="passRateColor(run)">{{ formatPassRate(run) }}</TableCell>
                <TableCell>{{ formatDuration(run) }}</TableCell>
                <TableCell class="text-muted-foreground">{{ formatDate(run) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <div v-else class="flex flex-col items-center justify-center py-16 text-sm text-muted-foreground gap-2">
          <LayoutDashboard class="h-10 w-10 text-muted-foreground/30 mb-1" />
          <span>No runs yet.</span>
          <span class="text-xs">Start one from the Runs page.</span>
        </div>
      </div>
    </template>

    <div v-else class="flex flex-col items-center justify-center py-16 text-sm text-muted-foreground gap-2">
      <LayoutDashboard class="h-10 w-10 text-muted-foreground/30 mb-1" />
      <span>No project selected.</span>
      <span class="text-xs">Add a project to get started.</span>
    </div>
  </div>
</template>
