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
import { Card, CardContent } from '@/components/ui/card'
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
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">Runs</h1>
      <div class="flex items-center gap-2">
        <Button variant="outline" :disabled="!canCompare" @click="goCompare">Compare</Button>
        <Button :disabled="!currentProjectName" @click="showNewRun = true">New Run</Button>
      </div>
    </div>

    <template v-if="currentProjectName">
      <template v-if="activeRuns.length > 0">
        <h3 class="text-base font-medium text-zinc-900">Active Runs</h3>
        <Card
          v-for="run in activeRuns"
          :key="run.id"
          class="cursor-pointer hover:shadow-md transition-shadow border-l-4 border-l-primary"
          @click="viewActiveRun(run)"
        >
          <CardContent class="py-3 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <Badge class="bg-blue-50 text-blue-600 border-0">Running</Badge>
              <code class="text-sm bg-zinc-100 px-1.5 py-0.5 rounded">{{ run.id.slice(0, 8) }}</code>
            </div>
            <span class="text-sm text-muted-foreground">Started {{ formatDate(run.started_at) }}</span>
          </CardContent>
        </Card>
      </template>

      <h3 class="text-base font-medium text-zinc-900">Run History</h3>
      <template v-if="recentRuns.length > 0">
        <Card v-for="run in recentRuns" :key="run.id" class="hover:shadow-sm transition-shadow">
          <CardContent class="py-3 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <input
                type="checkbox"
                :checked="isRunSelected(run.id)"
                :disabled="isRunDisabled(run.id)"
                class="h-4 w-4 rounded border-gray-300 text-primary accent-primary"
                @change="toggleRunSelection(run.id)"
                @click.stop
              />
              <div class="flex items-center gap-2 cursor-pointer" @click="viewRun(run)">
                <span class="text-sm font-medium text-zinc-900">{{ run.suite_name }}</span>
                <span class="text-sm text-muted-foreground">{{ run.agent_type }}</span>
              </div>
            </div>
            <div class="flex items-center gap-3 cursor-pointer" @click="viewRun(run)">
              <Badge class="bg-success-light text-success border-0">
                {{ ((run.summary?.overall_pass_rate ?? 0) * 100).toFixed(0) }}%
              </Badge>
              <span class="text-sm text-muted-foreground">{{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</span>
              <span class="text-sm text-muted-foreground">{{ formatDate(run.started_at) }}</span>
            </div>
          </CardContent>
        </Card>
      </template>
      <div v-else class="flex items-center justify-center py-12 text-sm text-muted-foreground">
        No runs yet
      </div>
    </template>

    <div v-else class="flex items-center justify-center py-12 text-sm text-muted-foreground">
      Select a project first
    </div>

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
