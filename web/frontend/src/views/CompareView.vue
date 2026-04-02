<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { ArrowLeft, ChevronRight } from 'lucide-vue-next'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useRoute, useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { compareRuns } from '@/api/compare'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { CompareResult, TaskComparison } from '@/types'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent])

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)

const result = ref<CompareResult | null>(null)
const loading = ref(true)
const statusFilter = ref<string>('all')
const expandedTasks = ref<Set<string>>(new Set())

onMounted(async () => {
  const runA = route.query.runA as string
  const runB = route.query.runB as string
  if (!runA || !runB || !currentProjectName.value) {
    toast.error('Missing run IDs')
    router.push({ name: 'runs' })
    return
  }
  try {
    result.value = await compareRuns(currentProjectName.value, runA, runB)
  } catch (e: unknown) {
    toast.error(`Failed to compare runs: ${e instanceof Error ? e.message : 'Unknown error'}`)
    router.push({ name: 'runs' })
  } finally {
    loading.value = false
  }
})

const chartOption = computed(() => {
  if (!result.value) return {}
  const s = result.value.summary
  return {
    tooltip: { trigger: 'axis' as const },
    legend: { data: ['Run A', 'Run B'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category' as const, data: ['Pass Rate', 'Avg Score', 'pass@k', 'pass^k'] },
    yAxis: { type: 'value' as const, max: 1.0 },
    series: [
      { name: 'Run A', type: 'bar' as const, data: [s.pass_rate.a, s.avg_score.a, s.avg_pass_at_k.a, s.avg_pass_power_k.a], itemStyle: { color: '#f97316' } },
      { name: 'Run B', type: 'bar' as const, data: [s.pass_rate.b, s.avg_score.b, s.avg_pass_at_k.b, s.avg_pass_power_k.b], itemStyle: { color: '#6366f1' } },
    ],
  }
})

interface MetricRow {
  metric: string
  a: string
  b: string
  diff: string
  direction: 'up' | 'down' | 'equal'
}

const metricRows = computed<MetricRow[]>(() => {
  if (!result.value) return []
  const s = result.value.summary
  function fmtPct(v: number): string { return (v * 100).toFixed(1) + '%' }
  function fmtNum(v: number): string { return v.toFixed(3) }
  function fmtDiffPct(v: number): string { return (v >= 0 ? '+' : '') + (v * 100).toFixed(1) + '%' }
  function fmtDiffNum(v: number): string { return (v >= 0 ? '+' : '') + v.toFixed(3) }
  function dir(v: number): 'up' | 'down' | 'equal' {
    if (v > 0.01) return 'up'
    if (v < -0.01) return 'down'
    return 'equal'
  }
  return [
    { metric: 'Pass Rate', a: fmtPct(s.pass_rate.a), b: fmtPct(s.pass_rate.b), diff: fmtDiffPct(s.pass_rate.diff), direction: dir(s.pass_rate.diff) },
    { metric: 'Avg Score', a: fmtNum(s.avg_score.a), b: fmtNum(s.avg_score.b), diff: fmtDiffNum(s.avg_score.diff), direction: dir(s.avg_score.diff) },
    { metric: 'pass@k', a: fmtNum(s.avg_pass_at_k.a), b: fmtNum(s.avg_pass_at_k.b), diff: fmtDiffNum(s.avg_pass_at_k.diff), direction: dir(s.avg_pass_at_k.diff) },
    { metric: 'pass^k', a: fmtNum(s.avg_pass_power_k.a), b: fmtNum(s.avg_pass_power_k.b), diff: fmtDiffNum(s.avg_pass_power_k.diff), direction: dir(s.avg_pass_power_k.diff) },
  ]
})

const filteredTasks = computed(() => {
  if (!result.value) return []
  if (statusFilter.value === 'all') return result.value.tasks
  return result.value.tasks.filter((t) => t.status === statusFilter.value)
})

function toggleTask(taskId: string) {
  if (expandedTasks.value.has(taskId)) {
    expandedTasks.value.delete(taskId)
  } else {
    expandedTasks.value.add(taskId)
  }
}

function statusVariant(status: string): string {
  if (status === 'improved') return 'bg-success-light text-success border-0'
  if (status === 'regressed') return 'bg-error-light text-error border-0'
  return ''
}

function diffColor(direction: string): string {
  if (direction === 'up') return 'text-success'
  if (direction === 'down') return 'text-error'
  return 'text-muted-foreground'
}

function formatDate(s: string): string {
  return s ? new Date(s).toLocaleString() : '-'
}
</script>

<template>
  <div v-if="loading" class="flex items-center justify-center py-24">
    <div class="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
  </div>

  <div v-else-if="result" class="space-y-6">
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="sm" @click="router.push({ name: 'runs' })">
        <ArrowLeft class="h-4 w-4 mr-1" /> Back
      </Button>
      <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">Run Comparison</h1>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-sm">Run A</CardTitle>
        </CardHeader>
        <CardContent class="space-y-1 text-sm">
          <p>ID: <code class="bg-zinc-100 px-1 py-0.5 rounded text-xs">{{ result.run_a.id.slice(0, 8) }}</code></p>
          <p>Suite: {{ result.run_a.suite_name }}</p>
          <p>Agent: {{ result.run_a.agent_type }}</p>
          <p class="text-muted-foreground">{{ formatDate(result.run_a.started_at) }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-sm">Run B</CardTitle>
        </CardHeader>
        <CardContent class="space-y-1 text-sm">
          <p>ID: <code class="bg-zinc-100 px-1 py-0.5 rounded text-xs">{{ result.run_b.id.slice(0, 8) }}</code></p>
          <p>Suite: {{ result.run_b.suite_name }}</p>
          <p>Agent: {{ result.run_b.agent_type }}</p>
          <p class="text-muted-foreground">{{ formatDate(result.run_b.started_at) }}</p>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm">Summary Metrics</CardTitle>
      </CardHeader>
      <CardContent>
        <VChart :option="chartOption" style="height: 300px" autoresize />
        <div class="mt-4 bg-white rounded-lg border border-gray-200">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Metric</TableHead>
                <TableHead class="text-right">Run A</TableHead>
                <TableHead class="text-right">Run B</TableHead>
                <TableHead class="text-right">Diff</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="row in metricRows" :key="row.metric">
                <TableCell>{{ row.metric }}</TableCell>
                <TableCell class="text-right">{{ row.a }}</TableCell>
                <TableCell class="text-right">{{ row.b }}</TableCell>
                <TableCell class="text-right" :class="diffColor(row.direction)">
                  {{ row.direction === 'up' ? '↑' : row.direction === 'down' ? '↓' : '=' }}
                  {{ row.diff }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-2 flex flex-row items-center justify-between">
        <CardTitle class="text-sm">Per-Task Comparison</CardTitle>
        <Select v-model="statusFilter">
          <SelectTrigger class="w-[140px] h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="improved">Improved</SelectItem>
            <SelectItem value="regressed">Regressed</SelectItem>
            <SelectItem value="unchanged">Unchanged</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        <div class="space-y-1">
          <div v-for="task in filteredTasks" :key="task.task_id" class="border border-gray-200 rounded-lg overflow-hidden">
            <button
              class="flex items-center gap-3 w-full px-4 py-2.5 text-sm hover:bg-zinc-50 transition-colors text-left"
              @click="toggleTask(task.task_id)"
            >
              <ChevronRight
                class="h-4 w-4 text-muted-foreground transition-transform flex-shrink-0"
                :class="{ 'rotate-90': expandedTasks.has(task.task_id) }"
              />
              <span class="font-medium text-zinc-900 flex-1">{{ task.task_id }}</span>
              <span class="text-muted-foreground">{{ task.score_a.toFixed(3) }}</span>
              <span class="text-muted-foreground">→</span>
              <span class="text-muted-foreground">{{ task.score_b.toFixed(3) }}</span>
              <span :class="diffColor(task.diff > 0.01 ? 'up' : task.diff < -0.01 ? 'down' : 'equal')">
                {{ (task.diff >= 0 ? '+' : '') + task.diff.toFixed(3) }}
              </span>
              <Badge :class="statusVariant(task.status)">{{ task.status }}</Badge>
            </button>
            <div v-if="expandedTasks.has(task.task_id)" class="border-t border-gray-200 p-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-sm font-medium text-zinc-900 mb-2">Run A</p>
                  <div v-if="task.trials_a?.length" class="space-y-2">
                    <div v-for="(trial, i) in task.trials_a" :key="i" class="border border-gray-200 rounded p-2 text-xs space-y-1">
                      <div class="flex items-center gap-2">
                        <span>Trial #{{ i }}</span>
                        <Badge :class="trial.status === 'passed' ? 'bg-success-light text-success border-0' : trial.status === 'failed' ? 'bg-warning-light text-warning border-0' : 'bg-error-light text-error border-0'" class="text-xs">
                          {{ trial.status }}
                        </Badge>
                        <span class="text-muted-foreground">Score: {{ trial.score.toFixed(3) }}</span>
                      </div>
                      <div v-for="g in (trial.grades || [])" :key="g.grader_type" class="pl-3 flex items-center gap-1.5">
                        <Badge :class="g.pass ? 'bg-success-light text-success border-0' : 'bg-error-light text-error border-0'" class="text-xs">{{ g.grader_type }}</Badge>
                        <span>{{ g.score.toFixed(3) }}</span>
                        <span v-if="g.reason" class="text-muted-foreground">— {{ g.reason }}</span>
                      </div>
                    </div>
                  </div>
                  <p v-else class="text-sm text-muted-foreground">No Run A trials</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-zinc-900 mb-2">Run B</p>
                  <div v-if="task.trials_b?.length" class="space-y-2">
                    <div v-for="(trial, i) in task.trials_b" :key="i" class="border border-gray-200 rounded p-2 text-xs space-y-1">
                      <div class="flex items-center gap-2">
                        <span>Trial #{{ i }}</span>
                        <Badge :class="trial.status === 'passed' ? 'bg-success-light text-success border-0' : trial.status === 'failed' ? 'bg-warning-light text-warning border-0' : 'bg-error-light text-error border-0'" class="text-xs">
                          {{ trial.status }}
                        </Badge>
                        <span class="text-muted-foreground">Score: {{ trial.score.toFixed(3) }}</span>
                      </div>
                      <div v-for="g in (trial.grades || [])" :key="g.grader_type" class="pl-3 flex items-center gap-1.5">
                        <Badge :class="g.pass ? 'bg-success-light text-success border-0' : 'bg-error-light text-error border-0'" class="text-xs">{{ g.grader_type }}</Badge>
                        <span>{{ g.score.toFixed(3) }}</span>
                        <span v-if="g.reason" class="text-muted-foreground">— {{ g.reason }}</span>
                      </div>
                    </div>
                  </div>
                  <p v-else class="text-sm text-muted-foreground">No Run B trials</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
