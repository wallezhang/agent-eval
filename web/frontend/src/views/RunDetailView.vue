<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { NSpace, NText, NButton, NCard, NTag, useMessage } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { cancelRun, getRunSSEUrl } from '@/api/runs'
import { useSSE } from '@/composables/useSSE'
import RunProgress from '@/components/RunProgress.vue'
import type { SSERunStarted, SSETrialCompleted, SSERunProgress, SSERunCompleted, SSERunError } from '@/types'

const props = defineProps<{ id: string }>()

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()
const message = useMessage()

const suiteName = ref('')
const isRunning = ref(true)
const progress = ref({ completed: 0, total: 0, pass_count: 0, fail_count: 0, error_count: 0 })
const logs = ref<string[]>([])
const errorMsg = ref('')

const sseUrl = computed(() =>
  currentProjectName.value ? getRunSSEUrl(currentProjectName.value, props.id) : '',
)

const { events, connect } = useSSE(sseUrl.value)

watch(events, (evts) => {
  for (const evt of evts) {
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
        message.success('Run completed')
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
  events.value = []
}, { deep: true })

onMounted(() => {
  if (sseUrl.value) connect()
})

async function handleCancel() {
  if (!currentProjectName.value) return
  try {
    await cancelRun(currentProjectName.value, props.id)
    message.info('Cancelling run...')
  } catch {
    message.error('Failed to cancel')
  }
}

function viewResults() {
  router.push({ name: 'result-detail', params: { id: props.id } })
}
</script>

<template>
  <NSpace vertical :size="16">
    <NSpace justify="space-between" align="center">
      <NSpace align="center">
        <NButton quaternary @click="router.push({ name: 'runs' })">← Back</NButton>
        <NText tag="h1" style="margin: 0">{{ suiteName || 'Run' }}</NText>
        <NText code depth="3">{{ id.slice(0, 8) }}</NText>
        <NTag v-if="isRunning" type="info">Running</NTag>
        <NTag v-else-if="errorMsg" type="error">Error</NTag>
        <NTag v-else type="success">Completed</NTag>
      </NSpace>
      <NSpace>
        <NButton v-if="isRunning" type="error" @click="handleCancel">Cancel</NButton>
        <NButton v-if="!isRunning && !errorMsg" type="primary" @click="viewResults">View Results</NButton>
      </NSpace>
    </NSpace>

    <RunProgress
      :completed="progress.completed"
      :total="progress.total"
      :pass-count="progress.pass_count"
      :fail-count="progress.fail_count"
      :error-count="progress.error_count"
    />

    <NCard title="Log" size="small">
      <div style="background: #1a1a2e; color: #e0e0e0; font-family: monospace; font-size: 13px; padding: 12px; border-radius: 4px; max-height: 400px; overflow-y: auto; white-space: pre-wrap;">
        <div v-for="(log, i) in logs" :key="i">{{ log }}</div>
        <div v-if="logs.length === 0" style="color: #666">Waiting for events...</div>
      </div>
    </NCard>
  </NSpace>
</template>
