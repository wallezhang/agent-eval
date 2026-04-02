<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { toast } from 'vue-sonner'
import { ArrowLeft, ChevronRight } from 'lucide-vue-next'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { cancelRun, getRunSSEUrl } from '@/api/runs'
import { useSSE } from '@/composables/useSSE'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import RunProgress from '@/components/RunProgress.vue'
import type { SSERunStarted, SSETrialCompleted, SSERunProgress, SSERunCompleted, SSERunError } from '@/types'

const props = defineProps<{ id: string }>()

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()

const suiteName = ref('')
const isRunning = ref(true)
const progress = ref({ completed: 0, total: 0, pass_count: 0, fail_count: 0, error_count: 0 })
const logs = ref<string[]>([])
const errorMsg = ref('')

const sseUrl = computed(() =>
  currentProjectName.value ? getRunSSEUrl(currentProjectName.value, props.id) : '',
)

const sse = useSSE(sseUrl.value)

let lastProcessedIndex = 0
watch(
  () => sse.events.value.length,
  (newLen) => {
    if (newLen <= lastProcessedIndex) return
    const newEvents = sse.events.value.slice(lastProcessedIndex)
    lastProcessedIndex = newLen
    for (const evt of newEvents) {
      handleSSEEvent(evt)
    }
  },
)

function handleSSEEvent(evt: { type: string; data: unknown }) {
  switch (evt.type) {
    case 'run_started': {
      const d = evt.data as SSERunStarted
      suiteName.value = d.suite
      progress.value.total = d.total_tasks
      logs.value.push(`Run started: ${d.suite} (${d.total_tasks} tasks)`)
      break
    }
    case 'trial_completed': {
      const d = evt.data as SSETrialCompleted
      logs.value.push(`Trial ${d.task_id}#${d.trial_index}: ${d.status} (score: ${d.score}, ${d.duration_ms}ms)`)
      break
    }
    case 'run_progress': {
      const d = evt.data as SSERunProgress
      progress.value = { ...d }
      break
    }
    case 'run_completed': {
      const d = evt.data as SSERunCompleted
      isRunning.value = false
      logs.value.push(`Run completed: ${d.run_id.slice(0, 8)}`)
      toast.success('Run completed')
      break
    }
    case 'run_error': {
      const d = evt.data as SSERunError
      isRunning.value = false
      errorMsg.value = d.message
      logs.value.push(`Error: ${d.message}`)
      break
    }
    case 'log': {
      const d = evt.data as { message: string }
      logs.value.push(d.message)
      break
    }
  }
}

onMounted(() => {
  if (sseUrl.value) {
    sse.connect()
  }
})

async function handleCancel() {
  if (!currentProjectName.value) return
  try {
    await cancelRun(currentProjectName.value, props.id)
    toast.info('Cancelling run...')
  } catch {
    toast.error('Failed to cancel')
  }
}

function viewResults() {
  router.push({ name: 'result-detail', params: { id: props.id } })
}

function logLineClass(log: string): string {
  if (log.startsWith('Error:')) return 'text-red-400'
  if (log.includes(': passed')) return 'text-emerald-400'
  if (log.includes(': failed')) return 'text-amber-400'
  if (log.includes('Run completed')) return 'text-emerald-400 font-semibold'
  if (log.includes('Run started')) return 'text-blue-400'
  return ''
}
</script>

<template>
  <div class="space-y-6">
    <!-- Navigation breadcrumb -->
    <nav class="flex items-center gap-1.5 text-sm text-muted-foreground">
      <button
        class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md hover:bg-zinc-200/60 hover:text-zinc-900 transition-all duration-200 font-medium"
        @click="router.push({ name: 'runs' })"
      >
        <ArrowLeft class="h-3.5 w-3.5" />
        Runs
      </button>
      <ChevronRight class="h-3.5 w-3.5 text-muted-foreground/50" />
      <span class="text-zinc-700 font-medium">Live Run</span>
    </nav>

    <!-- Page header -->
    <div class="flex items-start justify-between">
      <div class="space-y-1.5">
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-extrabold text-zinc-900 tracking-tight font-display">{{ suiteName || 'Run' }}</h1>
          <code class="text-xs bg-zinc-100 border border-zinc-200 px-2 py-0.5 rounded-md text-muted-foreground font-mono">{{ id.slice(0, 8) }}</code>
          <div v-if="isRunning" class="relative">
            <Badge class="bg-blue-50 text-blue-600 border-0 rounded-full">Running</Badge>
            <span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-blue-500 rounded-full animate-pulse-glow" />
          </div>
          <Badge v-else-if="errorMsg" class="bg-error-light text-error border-0 rounded-full">Error</Badge>
          <Badge v-else class="bg-success-light text-success border-0 rounded-full">Completed</Badge>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Button v-if="isRunning" variant="destructive" @click="handleCancel">Cancel</Button>
        <Button v-if="!isRunning && !errorMsg" @click="viewResults">View Results</Button>
      </div>
    </div>

    <RunProgress
      :completed="progress.completed"
      :total="progress.total"
      :pass-count="progress.pass_count"
      :fail-count="progress.fail_count"
      :error-count="progress.error_count"
    />

    <Card class="card-shadow overflow-hidden">
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-display font-semibold">Log</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="bg-zinc-950 text-zinc-400 font-mono text-xs p-4 rounded-lg max-h-[400px] overflow-y-auto styled-scrollbar whitespace-pre-wrap relative">
          <div class="absolute inset-0 pointer-events-none bg-[repeating-linear-gradient(0deg,transparent,transparent_2px,rgba(0,0,0,0.03)_2px,rgba(0,0,0,0.03)_4px)]" />
          <div v-for="(log, i) in logs" :key="i" :class="logLineClass(log)" class="relative">{{ log }}</div>
          <div v-if="logs.length === 0" class="text-zinc-600 relative">Waiting for events...</div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
