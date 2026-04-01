<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NSpace, NText, NButton, NTag, NCollapse, NCollapseItem, useMessage } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { getRun } from '@/api/runs'
import SummaryCards from '@/components/SummaryCards.vue'
import type { SummaryCard } from '@/components/SummaryCards.vue'
import TrialTable from '@/components/TrialTable.vue'
import type { EvalRun, TaskResult } from '@/types'

const props = defineProps<{ id: string }>()

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()
const message = useMessage()

const run = ref<EvalRun | null>(null)
const loading = ref(true)

async function load() {
  if (!currentProjectName.value) return
  loading.value = true
  try {
    run.value = await getRun(currentProjectName.value, props.id)
  } catch {
    message.error('Failed to load run results')
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

function taskStatusType(tr: TaskResult): 'success' | 'warning' | 'error' {
  if (tr.error_count > 0) return 'error'
  if (tr.fail_count > 0) return 'warning'
  return 'success'
}
</script>

<template>
  <NSpace vertical :size="16">
    <NSpace align="center">
      <NButton quaternary @click="router.push({ name: 'runs' })">← Back</NButton>
      <NText tag="h1" style="margin: 0">{{ run?.suite_name || 'Results' }}</NText>
      <NText v-if="run" code depth="3">{{ run.id.slice(0, 8) }}</NText>
    </NSpace>

    <template v-if="run">
      <NSpace :size="8">
        <NText depth="3">Agent: {{ run.agent_type }}</NText>
        <NText depth="3">Duration: {{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</NText>
        <NText depth="3">{{ new Date(run.started_at).toLocaleString() }}</NText>
      </NSpace>

      <SummaryCards :cards="summaryCards()" />

      <NText tag="h3">Task Results</NText>
      <NCollapse>
        <NCollapseItem v-for="tr in run.task_results" :key="tr.task.id" :name="tr.task.id">
          <template #header>
            <NSpace align="center" :size="12" style="width: 100%">
              <NText strong>{{ tr.task.name || tr.task.id }}</NText>
              <NTag :type="taskStatusType(tr)" size="small">
                {{ tr.pass_count }}P / {{ tr.fail_count }}F / {{ tr.error_count }}E
              </NTag>
              <NText depth="3">Avg: {{ tr.avg_score.toFixed(3) }}</NText>
              <NText v-if="tr.latency_p50_ms" depth="3">
                P50: {{ tr.latency_p50_ms }}ms P90: {{ tr.latency_p90_ms }}ms
              </NText>
            </NSpace>
          </template>
          <TrialTable :trials="tr.trials" />
        </NCollapseItem>
      </NCollapse>
    </template>
  </NSpace>
</template>
