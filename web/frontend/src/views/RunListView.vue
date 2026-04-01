<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import {
  NSpace, NText, NButton, NTag, NEmpty, NModal, NSelect, NCard, useMessage,
} from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { useRunStore } from '@/stores/run'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { startRun } from '@/api/runs'
import { listConfigs } from '@/api/configs'
import type { EvalRun, ActiveRun } from '@/types'

const projectStore = useProjectStore()
const runStore = useRunStore()
const { currentProjectName } = storeToRefs(projectStore)
const { activeRuns, recentRuns } = storeToRefs(runStore)
const router = useRouter()
const message = useMessage()

const showNewRun = ref(false)
const configs = ref<string[]>([])
const selectedConfig = ref<string | null>(null)
const starting = ref(false)

async function loadData() {
  if (!currentProjectName.value) return
  await runStore.refresh(currentProjectName.value)
  configs.value = await listConfigs(currentProjectName.value)
}

watch(currentProjectName, loadData, { immediate: true })

async function handleStartRun() {
  if (!currentProjectName.value || !selectedConfig.value) return
  starting.value = true
  try {
    const resp = await startRun(currentProjectName.value, selectedConfig.value)
    showNewRun.value = false
    starting.value = false
    message.success(`Run started: ${resp.run_id.slice(0, 8)}`, { duration: 3000 })
    router.push({ name: 'run-detail', params: { id: resp.run_id } })
  } catch (e: unknown) {
    message.error(`Failed: ${e instanceof Error ? e.message : 'Unknown error'}`)
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

const configOptions = computed(() =>
  configs.value.map((c) => ({ label: c, value: c })),
)
</script>

<template>
  <NSpace vertical :size="16">
    <NSpace justify="space-between" align="center">
      <NText tag="h1" style="margin: 0">Runs</NText>
      <NButton type="primary" :disabled="!currentProjectName" @click="showNewRun = true">New Run</NButton>
    </NSpace>

    <template v-if="currentProjectName">
      <template v-if="activeRuns.length > 0">
        <NText tag="h3">Active Runs</NText>
        <NCard v-for="run in activeRuns" :key="run.id" size="small" hoverable style="cursor: pointer; margin-bottom: 8px" @click="viewActiveRun(run)">
          <NSpace justify="space-between" align="center">
            <NSpace align="center">
              <NTag type="info" size="small">Running</NTag>
              <NText code>{{ run.id.slice(0, 8) }}</NText>
            </NSpace>
            <NText depth="3">Started {{ formatDate(run.started_at) }}</NText>
          </NSpace>
        </NCard>
      </template>

      <NText tag="h3">Run History</NText>
      <template v-if="recentRuns.length > 0">
        <NCard v-for="run in recentRuns" :key="run.id" size="small" hoverable style="cursor: pointer; margin-bottom: 8px" @click="viewRun(run)">
          <NSpace justify="space-between" align="center">
            <NSpace align="center">
              <NText strong>{{ run.suite_name }}</NText>
              <NText depth="3">{{ run.agent_type }}</NText>
            </NSpace>
            <NSpace align="center">
              <NTag type="success" size="small">{{ ((run.summary?.overall_pass_rate ?? 0) * 100).toFixed(0) }}%</NTag>
              <NText depth="3">{{ ((run.duration_ms ?? 0) / 1000).toFixed(1) }}s</NText>
              <NText depth="3">{{ formatDate(run.started_at) }}</NText>
            </NSpace>
          </NSpace>
        </NCard>
      </template>
      <NEmpty v-else description="No runs yet" />
    </template>

    <NEmpty v-else description="Select a project first" />

    <NModal v-model:show="showNewRun" title="Start New Run" preset="card" style="width: 400px">
      <NSpace vertical :size="12">
        <NSelect v-model:value="selectedConfig" :options="configOptions" placeholder="Select config file" />
        <NSpace justify="end">
          <NButton @click="showNewRun = false">Cancel</NButton>
          <NButton type="primary" :loading="starting" :disabled="!selectedConfig" @click="handleStartRun">Start</NButton>
        </NSpace>
      </NSpace>
    </NModal>
  </NSpace>
</template>
