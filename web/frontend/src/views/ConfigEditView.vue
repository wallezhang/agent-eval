<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  NSpace, NText, NButton, NTag, NCard, NList, NListItem, NAlert, useMessage,
} from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { getConfig, updateConfig, validateConfig } from '@/api/configs'
import { listAgentTypes, listGraderTypes } from '@/api/meta'
import YamlEditor from '@/components/YamlEditor.vue'
import type { ValidationResult } from '@/types'

const props = defineProps<{ filename: string }>()

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const router = useRouter()
const message = useMessage()

const content = ref('')
const originalContent = ref('')
const validation = ref<ValidationResult>({ valid: true, errors: [] })
const saving = ref(false)
const agentTypes = ref<string[]>([])
const graderTypes = ref<string[]>([])

const isDirty = computed(() => content.value !== originalContent.value)

async function load() {
  if (!currentProjectName.value) return
  try {
    const data = await getConfig(currentProjectName.value, props.filename)
    content.value = data
    originalContent.value = data
  } catch {
    message.error('Failed to load config')
    router.push({ name: 'configs' })
  }
}

async function handleSave() {
  if (!currentProjectName.value) return
  saving.value = true
  try {
    await updateConfig(currentProjectName.value, props.filename, content.value)
    originalContent.value = content.value
    message.success('Saved')
    await handleValidate()
  } catch {
    message.error('Failed to save')
  } finally {
    saving.value = false
  }
}

async function handleValidate() {
  if (!currentProjectName.value) return
  try {
    validation.value = await validateConfig(currentProjectName.value, props.filename)
  } catch {
    validation.value = { valid: false, errors: ['Validation request failed'] }
  }
}

let validateTimer: ReturnType<typeof setTimeout> | null = null
watch(content, () => {
  if (validateTimer) clearTimeout(validateTimer)
  validateTimer = setTimeout(handleValidate, 800)
})

const agentTemplate = `agent:\n  type: openai\n  config:\n    model: gpt-4o\n    api_key: \${OPENAI_API_KEY}\n    temperature: 0\n`
const taskTemplate = `- id: task-name\n  name: "Task description"\n  input:\n    prompt: "prompt text"\n  expected:\n    text: "expected output"\n  graders:\n    - type: exact_match\n`
const graderTemplate = `- type: llm\n  config:\n    provider: openai\n    model: gpt-4o\n    prompt: "Rate the response..."\n`

function insertSnippet(snippet: string) {
  content.value = content.value + '\n' + snippet
}

onMounted(async () => {
  await load()
  try {
    agentTypes.value = await listAgentTypes()
    graderTypes.value = await listGraderTypes()
  } catch { /* non-critical */ }
})
</script>

<template>
  <NSpace vertical :size="16">
    <NSpace justify="space-between" align="center">
      <NSpace align="center">
        <NButton quaternary @click="router.push({ name: 'configs' })">← Back</NButton>
        <NText strong>{{ filename }}</NText>
        <NTag v-if="validation.valid" type="success" size="small">Valid</NTag>
        <NTag v-else type="error" size="small">{{ validation.errors.length }} error(s)</NTag>
      </NSpace>
      <NButton type="primary" :loading="saving" :disabled="!isDirty" @click="handleSave">Save</NButton>
    </NSpace>

    <NAlert v-if="!validation.valid && validation.errors.length > 0" type="error" title="Validation Errors">
      <ul style="margin: 0; padding-left: 20px">
        <li v-for="(err, i) in validation.errors" :key="i">{{ err }}</li>
      </ul>
    </NAlert>

    <div style="display: flex; gap: 16px; height: calc(100vh - 200px)">
      <div style="flex: 1; min-width: 0">
        <YamlEditor v-model="content" />
      </div>
      <div style="width: 260px; flex-shrink: 0">
        <NSpace vertical :size="12">
          <NCard title="Quick Insert" size="small">
            <NSpace vertical :size="8">
              <NButton block size="small" @click="insertSnippet(agentTemplate)">Agent Template</NButton>
              <NButton block size="small" @click="insertSnippet(taskTemplate)">Task Template</NButton>
              <NButton block size="small" @click="insertSnippet(graderTemplate)">Grader Template</NButton>
            </NSpace>
          </NCard>
          <NCard title="Agent Types" size="small">
            <NList :show-divider="false" size="small">
              <NListItem v-for="t in agentTypes" :key="t"><NText code>{{ t }}</NText></NListItem>
            </NList>
          </NCard>
          <NCard title="Grader Types" size="small">
            <NList :show-divider="false" size="small">
              <NListItem v-for="t in graderTypes" :key="t"><NText code>{{ t }}</NText></NListItem>
            </NList>
          </NCard>
        </NSpace>
      </div>
    </div>
  </NSpace>
</template>
