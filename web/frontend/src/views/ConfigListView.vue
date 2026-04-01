<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  NSpace, NText, NButton, NPopconfirm, NEmpty, NInput, NModal, NForm, NFormItem, useMessage,
} from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { listConfigs, deleteConfig, createConfig } from '@/api/configs'
import { startRun } from '@/api/runs'

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()
const message = useMessage()

const configs = ref<string[]>([])
const loading = ref(false)
const showNewModal = ref(false)
const newFilename = ref('')

async function loadConfigs() {
  if (!currentProjectName.value) return
  loading.value = true
  try {
    configs.value = await listConfigs(currentProjectName.value)
  } catch {
    message.error('Failed to load configs')
  } finally {
    loading.value = false
  }
}

watch(currentProjectName, loadConfigs, { immediate: true })

function editConfig(filename: string) {
  router.push({ name: 'config-edit', params: { filename } })
}

async function handleDelete(filename: string) {
  if (!currentProjectName.value) return
  try {
    await deleteConfig(currentProjectName.value, filename)
    message.success(`Deleted ${filename}`)
    await loadConfigs()
  } catch {
    message.error(`Failed to delete ${filename}`)
  }
}

async function handleRun(filename: string) {
  if (!currentProjectName.value) return
  try {
    const resp = await startRun(currentProjectName.value, filename)
    message.success(`Run started: ${resp.run_id.slice(0, 8)}`)
    router.push({ name: 'run-detail', params: { id: resp.run_id } })
  } catch (e: unknown) {
    message.error(`Failed to start run: ${e instanceof Error ? e.message : 'Unknown error'}`)
  }
}

async function handleCreate() {
  if (!currentProjectName.value || !newFilename.value) return
  const filename = newFilename.value.endsWith('.yaml')
    ? newFilename.value
    : `${newFilename.value}.yaml`
  try {
    const template = `name: ${filename.replace('.yaml', '')}\nagent:\n  type: openai\n  config:\n    api_key: \${OPENAI_API_KEY}\n    model: gpt-4o\ntasks: []\n`
    await createConfig(currentProjectName.value, filename, template)
    message.success(`Created ${filename}`)
    showNewModal.value = false
    newFilename.value = ''
    router.push({ name: 'config-edit', params: { filename } })
  } catch {
    message.error('Failed to create config')
  }
}
</script>

<template>
  <NSpace vertical :size="16">
    <NSpace justify="space-between" align="center">
      <NText tag="h1" style="margin: 0">Configurations</NText>
      <NButton type="primary" :disabled="!currentProjectName" @click="showNewModal = true">
        New Config
      </NButton>
    </NSpace>

    <template v-if="currentProjectName">
      <div v-if="configs.length > 0">
        <div
          v-for="cfg in configs" :key="cfg"
          style="display: flex; justify-content: space-between; align-items: center; padding: 12px 0; border-bottom: 1px solid #eee"
        >
          <NText>{{ cfg }}</NText>
          <NSpace>
            <NButton size="small" @click="editConfig(cfg)">Edit</NButton>
            <NButton size="small" type="primary" @click="handleRun(cfg)">Run</NButton>
            <NPopconfirm @positive-click="handleDelete(cfg)">
              <template #trigger>
                <NButton size="small" type="error">Delete</NButton>
              </template>
              Delete {{ cfg }}?
            </NPopconfirm>
          </NSpace>
        </div>
      </div>
      <NEmpty v-else description="No config files. Create one to get started." />
    </template>

    <NEmpty v-else description="Select a project first" />

    <NModal v-model:show="showNewModal" title="New Config" preset="dialog" positive-text="Create" @positive-click="handleCreate">
      <NForm>
        <NFormItem label="Filename">
          <NInput v-model:value="newFilename" placeholder="eval.yaml" />
        </NFormItem>
      </NForm>
    </NModal>
  </NSpace>
</template>
