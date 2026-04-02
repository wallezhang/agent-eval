<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { ArrowLeft, ChevronRight } from 'lucide-vue-next'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { getRun } from '@/api/runs'
import { Badge } from '@/components/ui/badge'
import SummaryCards from '@/components/SummaryCards.vue'
import type { SummaryCard } from '@/components/SummaryCards.vue'
import TrialTable from '@/components/TrialTable.vue'
import type { EvalRun, TaskResult } from '@/types'

const props = defineProps<{ id: string }>()

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()

const run = ref<EvalRun | null>(null)
const loading = ref(true)
const expandedTasks = ref<Set<string>>(new Set())

async function load() {
  if (!currentProjectName.value) return
  loading.value = true
  try {
    run.value = await getRun(currentProjectName.value, props.id)
  } catch {
    toast.error('Failed to load run results')
    router.push({ name: 'runs' })
  } finally {
    loading.value = false
  }
}

onMounted(load)

function summaryCards(): SummaryCard[] {
  if (!run.value?.summary) return []
  const s = run.value.summary
  return [
    { label: 'Pass Rate', value: `${(s.overall_pass_rate * 100).toFixed(1)}`, suffix: '%' },
    { label: 'Avg Score', value: s.avg_score.toFixed(3) },
    { label: 'Total Trials', value: s.total_trials },
    { label: 'Cost', value: s.usage?.estimated_cost_usd ? `$${s.usage.estimated_cost_usd.toFixed(4)}` : 'N/A' },
  ]
}

function taskStatusVariant(tr: TaskResult): string {
  if (tr.error_count > 0) return 'bg-error-light text-error border-0 rounded-full shadow-sm'
  if (tr.fail_count > 0) return 'bg-warning-light text-warning border-0 rounded-full shadow-sm'
  return 'bg-success-light text-success border-0 rounded-full shadow-sm'
}

function toggleTask(taskId: string) {
  if (expandedTasks.value.has(taskId)) {
    expandedTasks.value.delete(taskId)
  } else {
    expandedTasks.value.add(taskId)
  }
}

function scoreColor(score: number): string {
  if (score >= 0.8) return 'text-success font-bold font-display'
  if (score >= 0.5) return 'text-warning font-bold font-display'
  return 'text-error font-bold font-display'
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
      <span class="text-zinc-700 font-medium">Results</span>
    </nav>

    <!-- Page header -->
    <div class="flex items-start justify-between">
      <div class="space-y-1.5">
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-extrabold text-zinc-900 tracking-tight font-display">{{ run?.suite_name || 'Results' }}</h1>
          <code v-if="run" class="text-xs bg-zinc-100 border border-zinc-200 px-2 py-0.5 rounded-md text-muted-foreground font-mono">{{ run.id.slice(0, 8) }}</code>
        </div>
        <div v-if="run" class="flex items-center gap-3 text-sm text-muted-foreground">
          <span class="inline-flex items-center gap-1.5"><span class="w-1.5 h-1.5 rounded-full bg-primary" /> {{ run.agent_type }}</span>
          <span class="text-zinc-300">|</span>
          <span>{{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</span>
          <span class="text-zinc-300">|</span>
          <span>{{ new Date(run.started_at).toLocaleString() }}</span>
        </div>
      </div>
    </div>

    <template v-if="run">

      <SummaryCards :cards="summaryCards()" />

      <div>
        <h3 class="text-base font-semibold text-zinc-900 mb-3 font-display">Task Results</h3>
        <div class="space-y-1">
          <div
            v-for="(tr, index) in run.task_results"
            :key="tr.task.id"
            class="border border-gray-200 rounded-lg overflow-hidden transition-all duration-200 hover:border-primary/30"
            :class="index % 2 === 1 ? 'bg-zinc-50/30' : 'bg-white'"
          >
            <button
              class="flex items-center gap-3 w-full px-4 py-3 text-sm hover:bg-primary-light/20 transition-all duration-200 text-left group"
              @click="toggleTask(tr.task.id)"
            >
              <ChevronRight
                class="h-4 w-4 text-muted-foreground transition-transform flex-shrink-0"
                :class="{ 'rotate-90': expandedTasks.has(tr.task.id) }"
              />
              <span class="font-medium text-zinc-900 font-display">{{ tr.task.name || tr.task.id }}</span>
              <Badge :class="taskStatusVariant(tr)">
                {{ tr.pass_count }}P / {{ tr.fail_count }}F / {{ tr.error_count }}E
              </Badge>
              <span :class="scoreColor(tr.avg_score)">{{ tr.avg_score.toFixed(3) }}</span>
              <span v-if="tr.latency_p50_ms" class="text-muted-foreground text-xs">
                P50: {{ tr.latency_p50_ms }}ms P90: {{ tr.latency_p90_ms }}ms
              </span>
            </button>
            <div v-if="expandedTasks.has(tr.task.id)" class="border-t border-gray-200 p-4 border-l-2 border-l-primary/20 ml-2">
              <TrialTable :trials="tr.trials" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
