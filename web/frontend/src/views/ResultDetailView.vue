<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { ArrowLeft, ChevronRight } from 'lucide-vue-next'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { getRun } from '@/api/runs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
  if (tr.error_count > 0) return 'bg-error-light text-error border-0'
  if (tr.fail_count > 0) return 'bg-warning-light text-warning border-0'
  return 'bg-success-light text-success border-0'
}

function toggleTask(taskId: string) {
  if (expandedTasks.value.has(taskId)) {
    expandedTasks.value.delete(taskId)
  } else {
    expandedTasks.value.add(taskId)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="sm" @click="router.push({ name: 'runs' })">
        <ArrowLeft class="h-4 w-4 mr-1" /> Back
      </Button>
      <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">{{ run?.suite_name || 'Results' }}</h1>
      <code v-if="run" class="text-sm bg-zinc-100 px-1.5 py-0.5 rounded text-muted-foreground">{{ run.id.slice(0, 8) }}</code>
    </div>

    <template v-if="run">
      <div class="flex items-center gap-4 text-sm text-muted-foreground">
        <span>Agent: {{ run.agent_type }}</span>
        <span>Duration: {{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</span>
        <span>{{ new Date(run.started_at).toLocaleString() }}</span>
      </div>

      <SummaryCards :cards="summaryCards()" />

      <div>
        <h3 class="text-base font-medium text-zinc-900 mb-3">Task Results</h3>
        <div class="space-y-1">
          <div v-for="tr in run.task_results" :key="tr.task.id" class="border border-gray-200 rounded-lg overflow-hidden">
            <button
              class="flex items-center gap-3 w-full px-4 py-3 text-sm hover:bg-zinc-50 transition-colors text-left"
              @click="toggleTask(tr.task.id)"
            >
              <ChevronRight
                class="h-4 w-4 text-muted-foreground transition-transform flex-shrink-0"
                :class="{ 'rotate-90': expandedTasks.has(tr.task.id) }"
              />
              <span class="font-medium text-zinc-900">{{ tr.task.name || tr.task.id }}</span>
              <Badge :class="taskStatusVariant(tr)">
                {{ tr.pass_count }}P / {{ tr.fail_count }}F / {{ tr.error_count }}E
              </Badge>
              <span class="text-muted-foreground">Avg: {{ tr.avg_score.toFixed(3) }}</span>
              <span v-if="tr.latency_p50_ms" class="text-muted-foreground">
                P50: {{ tr.latency_p50_ms }}ms P90: {{ tr.latency_p90_ms }}ms
              </span>
            </button>
            <div v-if="expandedTasks.has(tr.task.id)" class="border-t border-gray-200 p-4">
              <TrialTable :trials="tr.trials" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
