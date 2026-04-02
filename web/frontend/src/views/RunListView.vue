<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { toast } from 'vue-sonner'
import { useProjectStore } from '@/stores/project'
import { useRunStore } from '@/stores/run'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { startRun } from '@/api/runs'
import { listConfigs } from '@/api/configs'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Play, Plus, GitCompareArrows, Timer, Calendar, Zap, ArrowRight } from 'lucide-vue-next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { EvalRun, ActiveRun } from '@/types'

const projectStore = useProjectStore()
const runStore = useRunStore()
const { currentProjectName } = storeToRefs(projectStore)
const { activeRuns, recentRuns } = storeToRefs(runStore)
const router = useRouter()

const showNewRun = ref(false)
const configs = ref<string[]>([])
const selectedConfig = ref<string | null>(null)
const starting = ref(false)
const selectedRunIds = ref<string[]>([])

const canCompare = computed(() => selectedRunIds.value.length === 2)

async function loadData() {
  if (!currentProjectName.value) return
  selectedRunIds.value = []
  await runStore.refresh(currentProjectName.value)
  configs.value = await listConfigs(currentProjectName.value)
}

watch(currentProjectName, loadData, { immediate: true })

function toggleRunSelection(runId: string) {
  const idx = selectedRunIds.value.indexOf(runId)
  if (idx >= 0) {
    selectedRunIds.value.splice(idx, 1)
  } else if (selectedRunIds.value.length < 2) {
    selectedRunIds.value.push(runId)
  }
}

function isRunSelected(runId: string): boolean {
  return selectedRunIds.value.includes(runId)
}

function isRunDisabled(runId: string): boolean {
  return selectedRunIds.value.length >= 2 && !selectedRunIds.value.includes(runId)
}

function goCompare() {
  if (!canCompare.value) return
  router.push({
    name: 'compare',
    query: { runA: selectedRunIds.value[0], runB: selectedRunIds.value[1] },
  })
}

async function handleStartRun() {
  if (!currentProjectName.value || !selectedConfig.value) return
  starting.value = true
  try {
    const resp = await startRun(currentProjectName.value, selectedConfig.value)
    showNewRun.value = false
    starting.value = false
    toast.success(`Run started: ${resp.run_id.slice(0, 8)}`)
    router.push({ name: 'run-detail', params: { id: resp.run_id } })
  } catch (e: unknown) {
    toast.error(`Failed: ${e instanceof Error ? e.message : 'Unknown error'}`)
    starting.value = false
  }
}

function viewRun(run: EvalRun) {
  router.push({ name: 'result-detail', params: { id: run.id } })
}

function viewActiveRun(run: ActiveRun) {
  router.push({ name: 'run-detail', params: { id: run.id } })
}

function formatDate(s: string): string {
  return s ? new Date(s).toLocaleString() : '-'
}

function formatRelativeDate(s: string): string {
  if (!s) return '-'
  const d = new Date(s)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  const diffDay = Math.floor(diffHr / 24)
  if (diffDay < 7) return `${diffDay}d ago`
  return d.toLocaleDateString()
}

function passRateColor(rate: number): string {
  if (rate >= 0.8) return 'text-success'
  if (rate >= 0.5) return 'text-warning'
  return 'text-error'
}

function passRateBg(rate: number): string {
  if (rate >= 0.8) return 'bg-success/10 text-success'
  if (rate >= 0.5) return 'bg-warning/10 text-warning'
  return 'bg-error/10 text-error'
}

function selectionLabel(runId: string): string {
  const idx = selectedRunIds.value.indexOf(runId)
  if (idx === 0) return 'A'
  if (idx === 1) return 'B'
  return ''
}
</script>

<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-extrabold text-zinc-900 tracking-tight font-display">Runs</h1>
        <p class="text-sm text-muted-foreground mt-1">Manage and compare evaluation runs</p>
      </div>
      <button
        :disabled="!currentProjectName"
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-gradient-to-r from-primary to-amber-500 text-white font-display font-semibold text-sm shadow-sm hover:shadow-md hover:brightness-105 transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:shadow-sm"
        @click="showNewRun = true"
      >
        <Plus class="h-4 w-4" />
        New Run
      </button>
    </div>

    <template v-if="currentProjectName">
      <!-- Active runs -->
      <template v-if="activeRuns.length > 0">
        <div class="space-y-3">
          <h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-widest font-display">Active</h3>
          <div
            v-for="run in activeRuns"
            :key="run.id"
            class="group relative bg-white rounded-xl border border-blue-200/60 card-shadow cursor-pointer overflow-hidden transition-all duration-200 hover:border-blue-300 hover:shadow-md"
            @click="viewActiveRun(run)"
          >
            <!-- Animated top accent -->
            <div class="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-400 via-primary to-amber-400 animate-gradient" />
            <div class="px-5 py-4 flex items-center justify-between">
              <div class="flex items-center gap-4">
                <div class="relative">
                  <div class="w-9 h-9 rounded-lg bg-blue-50 flex items-center justify-center">
                    <Zap class="h-4 w-4 text-blue-500" />
                  </div>
                  <span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-blue-500 rounded-full animate-pulse-glow" />
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <code class="text-sm font-mono font-semibold text-zinc-800">{{ run.id.slice(0, 8) }}</code>
                    <Badge class="bg-blue-50 text-blue-600 border-0 rounded-full text-xs">Running</Badge>
                  </div>
                  <p class="text-xs text-muted-foreground mt-0.5">Started {{ formatRelativeDate(run.started_at) }}</p>
                </div>
              </div>
              <ArrowRight class="h-4 w-4 text-muted-foreground/40 group-hover:text-primary group-hover:translate-x-0.5 transition-all duration-200" />
            </div>
          </div>
        </div>
      </template>

      <!-- Compare toolbar (contextual) -->
      <div
        v-if="selectedRunIds.length > 0"
        class="sticky top-0 z-10 flex items-center justify-between bg-white/90 backdrop-blur-md rounded-xl border border-primary/20 px-5 py-3 card-shadow animate-fade-in-up"
      >
        <div class="flex items-center gap-3 text-sm">
          <GitCompareArrows class="h-4 w-4 text-primary" />
          <span class="text-zinc-600">
            <span class="font-semibold text-zinc-900">{{ selectedRunIds.length }}</span> of 2 selected for comparison
          </span>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="text-sm text-muted-foreground hover:text-zinc-900 px-3 py-1.5 rounded-md hover:bg-zinc-100 transition-colors duration-200"
            @click="selectedRunIds = []"
          >
            Clear
          </button>
          <button
            :disabled="!canCompare"
            class="inline-flex items-center gap-1.5 text-sm font-semibold px-4 py-1.5 rounded-lg transition-all duration-200"
            :class="canCompare
              ? 'bg-primary text-white hover:bg-primary-hover shadow-sm'
              : 'bg-zinc-100 text-zinc-400 cursor-not-allowed'"
            @click="goCompare"
          >
            <GitCompareArrows class="h-3.5 w-3.5" />
            Compare
          </button>
        </div>
      </div>

      <!-- Run history table -->
      <div class="space-y-3">
        <h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-widest font-display">History</h3>
        <div v-if="recentRuns.length > 0" class="bg-white rounded-xl border border-gray-200 card-shadow overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow class="bg-zinc-50/60 hover:bg-zinc-50/60">
                <TableHead class="w-10" />
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Suite</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Agent</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">Pass Rate</TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">
                  <span class="inline-flex items-center gap-1"><Timer class="h-3 w-3" /> Duration</span>
                </TableHead>
                <TableHead class="font-display font-semibold text-xs uppercase tracking-wider">
                  <span class="inline-flex items-center gap-1"><Calendar class="h-3 w-3" /> Date</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="(run, index) in recentRuns"
                :key="run.id"
                class="group transition-colors duration-150"
                :class="[
                  isRunSelected(run.id) ? 'bg-primary-light/40' : (index % 2 === 1 ? 'bg-zinc-50/30' : ''),
                  isRunDisabled(run.id) ? 'opacity-40' : 'cursor-pointer hover:bg-primary-light/20',
                ]"
                @click="viewRun(run)"
              >
                <!-- Selection checkbox -->
                <TableCell class="w-10 pr-0" @click.stop>
                  <div class="relative flex items-center justify-center">
                    <input
                      type="checkbox"
                      :checked="isRunSelected(run.id)"
                      :disabled="isRunDisabled(run.id)"
                      class="h-4 w-4 rounded border-gray-300 text-primary accent-primary cursor-pointer disabled:cursor-not-allowed"
                      @change="toggleRunSelection(run.id)"
                    />
                    <!-- A/B label overlay -->
                    <span
                      v-if="selectionLabel(run.id)"
                      class="absolute -top-1.5 -right-1 w-4 h-4 rounded-full text-[10px] font-bold flex items-center justify-center"
                      :class="selectionLabel(run.id) === 'A' ? 'bg-primary text-white' : 'bg-indigo-500 text-white'"
                    >
                      {{ selectionLabel(run.id) }}
                    </span>
                  </div>
                </TableCell>

                <!-- Suite name -->
                <TableCell>
                  <div class="flex items-center gap-2.5">
                    <div class="w-0.5 h-5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-primary" />
                    <div>
                      <p class="font-medium text-zinc-900 font-display text-sm">{{ run.suite_name }}</p>
                      <code class="text-[11px] text-muted-foreground font-mono">{{ run.id.slice(0, 8) }}</code>
                    </div>
                  </div>
                </TableCell>

                <!-- Agent -->
                <TableCell>
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md bg-zinc-100 text-xs font-medium text-zinc-600">{{ run.agent_type }}</span>
                </TableCell>

                <!-- Pass rate -->
                <TableCell>
                  <div class="flex items-center gap-2">
                    <div class="w-16 h-1.5 bg-zinc-100 rounded-full overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-500"
                        :class="(run.summary?.overall_pass_rate ?? 0) >= 0.8 ? 'bg-success' : (run.summary?.overall_pass_rate ?? 0) >= 0.5 ? 'bg-warning' : 'bg-error'"
                        :style="{ width: ((run.summary?.overall_pass_rate ?? 0) * 100) + '%' }"
                      />
                    </div>
                    <span class="text-sm font-semibold font-display tabular-nums" :class="passRateColor(run.summary?.overall_pass_rate ?? 0)">
                      {{ ((run.summary?.overall_pass_rate ?? 0) * 100).toFixed(0) }}%
                    </span>
                  </div>
                </TableCell>

                <!-- Duration -->
                <TableCell>
                  <span class="text-sm text-zinc-600 tabular-nums">{{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</span>
                </TableCell>

                <!-- Date -->
                <TableCell>
                  <span class="text-sm text-muted-foreground" :title="formatDate(run.started_at)">{{ formatRelativeDate(run.started_at) }}</span>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <div v-else class="flex flex-col items-center justify-center py-20 text-sm text-muted-foreground gap-3">
          <div class="w-14 h-14 rounded-2xl bg-zinc-100 flex items-center justify-center">
            <Play class="h-6 w-6 text-muted-foreground/40" />
          </div>
          <div class="text-center">
            <p class="font-medium text-zinc-600">No runs yet</p>
            <p class="text-xs mt-0.5">Start a new evaluation run to see results here.</p>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="flex flex-col items-center justify-center py-20 text-sm text-muted-foreground gap-3">
      <div class="w-14 h-14 rounded-2xl bg-zinc-100 flex items-center justify-center">
        <Play class="h-6 w-6 text-muted-foreground/40" />
      </div>
      <div class="text-center">
        <p class="font-medium text-zinc-600">Select a project first</p>
        <p class="text-xs mt-0.5">Choose a project from the sidebar to view runs.</p>
      </div>
    </div>

    <!-- New Run Dialog -->
    <Dialog v-model:open="showNewRun">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Start New Run</DialogTitle>
        </DialogHeader>
        <div class="py-4">
          <Select v-model="selectedConfig">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Select config file" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="c in configs" :key="c" :value="c">{{ c }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button :disabled="!selectedConfig || starting" @click="handleStartRun">
            {{ starting ? 'Starting...' : 'Start' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
