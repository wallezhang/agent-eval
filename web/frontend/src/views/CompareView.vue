<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import {
  NSpace, NText, NCard, NGrid, NGridItem, NTag, NDataTable, NButton,
  NSelect, NSpin, useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useRoute, useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { compareRuns } from '@/api/compare'
import type { CompareResult, TaskComparison, CompareTrialDetail, CompareGradeDetail } from '@/types'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent])

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const message = useMessage()

const result = ref<CompareResult | null>(null)
const loading = ref(true)
const statusFilter = ref<string>('all')

const statusFilterOptions = [
  { label: 'All', value: 'all' },
  { label: 'Improved', value: 'improved' },
  { label: 'Regressed', value: 'regressed' },
  { label: 'Unchanged', value: 'unchanged' },
]

onMounted(async () => {
  const runA = route.query.runA as string
  const runB = route.query.runB as string
  if (!runA || !runB || !currentProjectName.value) {
    message.error('Missing run IDs')
    router.push({ name: 'runs' })
    return
  }
  try {
    result.value = await compareRuns(currentProjectName.value, runA, runB)
  } catch (e: unknown) {
    message.error(`Failed to compare runs: ${e instanceof Error ? e.message : 'Unknown error'}`)
    router.push({ name: 'runs' })
  } finally {
    loading.value = false
  }
})

// Chart options
const chartOption = computed(() => {
  if (!result.value) return {}
  const s = result.value.summary
  return {
    tooltip: { trigger: 'axis' as const },
    legend: { data: ['Run A', 'Run B'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category' as const,
      data: ['Pass Rate', 'Avg Score', 'pass@k', 'pass^k'],
    },
    yAxis: { type: 'value' as const, max: 1.0 },
    series: [
      {
        name: 'Run A',
        type: 'bar' as const,
        data: [s.pass_rate.a, s.avg_score.a, s.avg_pass_at_k.a, s.avg_pass_power_k.a],
        itemStyle: { color: '#5B8FF9' },
      },
      {
        name: 'Run B',
        type: 'bar' as const,
        data: [s.pass_rate.b, s.avg_score.b, s.avg_pass_at_k.b, s.avg_pass_power_k.b],
        itemStyle: { color: '#5AD8A6' },
      },
    ],
  }
})

// Metrics table
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

// Task comparison table
const filteredTasks = computed(() => {
  if (!result.value) return []
  if (statusFilter.value === 'all') return result.value.tasks
  return result.value.tasks.filter((t) => t.status === statusFilter.value)
})

const taskColumns: DataTableColumns<TaskComparison> = [
  { title: 'Task ID', key: 'task_id', sorter: (a, b) => a.task_id.localeCompare(b.task_id) },
  { title: 'Score A', key: 'score_a', sorter: (a, b) => a.score_a - b.score_a, render: (row) => row.score_a.toFixed(3) },
  { title: 'Score B', key: 'score_b', sorter: (a, b) => a.score_b - b.score_b, render: (row) => row.score_b.toFixed(3) },
  { title: 'Diff', key: 'diff', sorter: (a, b) => a.diff - b.diff, render: (row) => (row.diff >= 0 ? '+' : '') + row.diff.toFixed(3) },
  {
    title: 'Status',
    key: 'status',
    render: (row) => {
      const typeMap: Record<string, 'success' | 'error' | 'default'> = {
        improved: 'success',
        regressed: 'error',
        unchanged: 'default',
      }
      return h(NTag, { type: typeMap[row.status] || 'default', size: 'small' }, { default: () => row.status })
    },
  },
]

function renderTrialTable(trials: CompareTrialDetail[], label: string) {
  if (!trials || trials.length === 0) {
    return h(NText, { depth: 3 }, { default: () => `No ${label} trials` })
  }
  return h('div', { style: 'margin-bottom: 12px' }, [
    h(NText, { strong: true, style: 'margin-bottom: 4px; display: block' }, { default: () => label }),
    ...trials.map((trial, i) =>
      h(NCard, { size: 'small', style: 'margin-bottom: 4px' }, {
        default: () => h(NSpace, { vertical: true, size: 4 }, {
          default: () => [
            h(NSpace, { align: 'center', size: 8 }, {
              default: () => [
                h(NText, {}, { default: () => `Trial #${i}` }),
                h(NTag, {
                  type: trial.status === 'passed' ? 'success' : trial.status === 'failed' ? 'warning' : 'error',
                  size: 'small',
                }, { default: () => trial.status }),
                h(NText, { depth: 3 }, { default: () => `Score: ${trial.score.toFixed(3)}` }),
              ],
            }),
            ...(trial.grades || []).map((g: CompareGradeDetail) =>
              h(NSpace, { align: 'center', size: 8, style: 'padding-left: 16px' }, {
                default: () => [
                  h(NTag, { type: g.pass ? 'success' : 'error', size: 'small' }, { default: () => g.grader_type }),
                  h(NText, {}, { default: () => `Score: ${g.score.toFixed(3)}` }),
                  g.reason ? h(NText, { depth: 3 }, { default: () => `— ${g.reason}` }) : null,
                ],
              }),
            ),
          ],
        }),
      }),
    ),
  ])
}

function renderExpandedRow(row: TaskComparison) {
  return h(NGrid, { cols: 2, xGap: 16 }, {
    default: () => [
      h(NGridItem, {}, { default: () => renderTrialTable(row.trials_a, 'Run A') }),
      h(NGridItem, {}, { default: () => renderTrialTable(row.trials_b, 'Run B') }),
    ],
  })
}

function formatDate(s: string): string {
  return s ? new Date(s).toLocaleString() : '-'
}
</script>

<template>
  <NSpin :show="loading">
    <NSpace vertical :size="16" v-if="result">
      <!-- Header -->
      <NSpace align="center">
        <NButton quaternary @click="router.push({ name: 'runs' })">← Back</NButton>
        <NText tag="h1" style="margin: 0">Run Comparison</NText>
      </NSpace>

      <!-- Run Meta Cards -->
      <NGrid :cols="2" :x-gap="16">
        <NGridItem>
          <NCard title="Run A" size="small">
            <NSpace vertical :size="4">
              <NText>ID: <NText code>{{ result.run_a.id.slice(0, 8) }}</NText></NText>
              <NText>Suite: {{ result.run_a.suite_name }}</NText>
              <NText>Agent: {{ result.run_a.agent_type }}</NText>
              <NText depth="3">{{ formatDate(result.run_a.started_at) }}</NText>
            </NSpace>
          </NCard>
        </NGridItem>
        <NGridItem>
          <NCard title="Run B" size="small">
            <NSpace vertical :size="4">
              <NText>ID: <NText code>{{ result.run_b.id.slice(0, 8) }}</NText></NText>
              <NText>Suite: {{ result.run_b.suite_name }}</NText>
              <NText>Agent: {{ result.run_b.agent_type }}</NText>
              <NText depth="3">{{ formatDate(result.run_b.started_at) }}</NText>
            </NSpace>
          </NCard>
        </NGridItem>
      </NGrid>

      <!-- Summary Chart -->
      <NCard title="Summary Metrics" size="small">
        <VChart :option="chartOption" style="height: 300px" autoresize />
        <table style="width: 100%; margin-top: 12px; border-collapse: collapse;">
          <thead>
            <tr style="border-bottom: 1px solid #eee;">
              <th style="text-align: left; padding: 4px 8px;">Metric</th>
              <th style="text-align: right; padding: 4px 8px;">Run A</th>
              <th style="text-align: right; padding: 4px 8px;">Run B</th>
              <th style="text-align: right; padding: 4px 8px;">Diff</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in metricRows" :key="row.metric" style="border-bottom: 1px solid #f5f5f5;">
              <td style="padding: 4px 8px;">{{ row.metric }}</td>
              <td style="text-align: right; padding: 4px 8px;">{{ row.a }}</td>
              <td style="text-align: right; padding: 4px 8px;">{{ row.b }}</td>
              <td style="text-align: right; padding: 4px 8px;">
                <span :style="{ color: row.direction === 'up' ? '#18a058' : row.direction === 'down' ? '#d03050' : '#999' }">
                  {{ row.direction === 'up' ? '↑' : row.direction === 'down' ? '↓' : '=' }}
                  {{ row.diff }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </NCard>

      <!-- Task Comparison Table -->
      <NCard title="Per-Task Comparison" size="small">
        <template #header-extra>
          <NSelect
            v-model:value="statusFilter"
            :options="statusFilterOptions"
            size="small"
            style="width: 140px"
          />
        </template>
        <NDataTable
          :columns="taskColumns"
          :data="filteredTasks"
          :row-key="(row: TaskComparison) => row.task_id"
          :default-expand-all="false"
          :render-expand="renderExpandedRow"
          size="small"
        />
      </NCard>
    </NSpace>
  </NSpin>
</template>
