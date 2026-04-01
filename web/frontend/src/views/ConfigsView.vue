<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import {
  NSpace,
  NText,
  NButton,
  NTag,
  NCard,
  NList,
  NListItem,
  NAlert,
  NEmpty,
  useMessage,
} from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { getConfig, updateConfig, createConfig, validateConfig } from '@/api/configs'
import { listFileTree, createDir } from '@/api/files'
import { listAgentTypes, listGraderTypes } from '@/api/meta'
import FileTree from '@/components/FileTree.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import type { FileNode, ValidationResult } from '@/types'

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)
const message = useMessage()

// File tree state
const tree = ref<FileNode[]>([])
const selectedFile = ref<string | null>(null)

// Editor state
const content = ref('')
const originalContent = ref('')
const validation = ref<ValidationResult>({ valid: true, errors: [] })
const saving = ref(false)
const agentTypes = ref<string[]>([])
const graderTypes = ref<string[]>([])

const isDirty = ref(false)
const isLoadingFile = ref(false)
watch(content, (val) => {
  isDirty.value = val !== originalContent.value
})

// Load file tree
async function loadTree() {
  if (!currentProjectName.value) return
  try {
    tree.value = await listFileTree(currentProjectName.value)
  } catch {
    message.error('Failed to load file tree')
  }
}

// Load file content
async function loadFile(path: string) {
  if (!currentProjectName.value) return
  try {
    const data = await getConfig(currentProjectName.value, path)
    // Suppress content watcher validation during file load
    isLoadingFile.value = true
    selectedFile.value = path
    content.value = data
    originalContent.value = data
    validation.value = { valid: true, errors: [] }
    // Single explicit validation for the newly loaded file
    await handleValidate()
  } catch {
    message.error(`Failed to load ${path}`)
  } finally {
    isLoadingFile.value = false
  }
}

// Save file
async function handleSave() {
  if (!currentProjectName.value || !selectedFile.value) return
  saving.value = true
  try {
    await updateConfig(currentProjectName.value, selectedFile.value, content.value)
    originalContent.value = content.value
    isDirty.value = false
    message.success('Saved')
    await handleValidate()
  } catch {
    message.error('Failed to save')
  } finally {
    saving.value = false
  }
}

// Validate file
async function handleValidate() {
  if (!currentProjectName.value || !selectedFile.value) return
  try {
    validation.value = await validateConfig(currentProjectName.value, selectedFile.value)
  } catch {
    validation.value = { valid: false, errors: ['Validation request failed'] }
  }
}

let validateTimer: ReturnType<typeof setTimeout> | null = null
watch(content, () => {
  if (validateTimer) clearTimeout(validateTimer)
  // Skip debounced validation when loadFile is setting content
  if (isLoadingFile.value) return
  if (selectedFile.value) {
    validateTimer = setTimeout(handleValidate, 800)
  }
})

// Select file from tree
function handleSelectFile(path: string) {
  if (isDirty.value) {
    // Could prompt to save, for now just switch
  }
  loadFile(path)
}

// Create file
async function handleCreateFile(dir: string, filename: string) {
  if (!currentProjectName.value) return
  const fullPath = dir ? `${dir}/${filename}` : filename
  const template = `# ${filename}\n`
  try {
    await createConfig(currentProjectName.value, fullPath, template)
    message.success(`Created ${fullPath}`)
    await loadTree()
    loadFile(fullPath)
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : 'Failed to create file')
  }
}

// Create directory
async function handleCreateDir(path: string) {
  if (!currentProjectName.value) return
  try {
    await createDir(currentProjectName.value, path)
    message.success(`Created folder ${path}`)
    await loadTree()
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : 'Failed to create folder')
  }
}

// Quick insert templates
const agentTemplate = `agent:\n  type: openai\n  config:\n    model: gpt-4o\n    api_key: \${OPENAI_API_KEY}\n    temperature: 0\n`
const taskTemplate = `- id: task-name\n  name: "Task description"\n  input:\n    prompt: "prompt text"\n  expected:\n    text: "expected output"\n  graders:\n    - type: exact_match\n`
const graderTemplate = `- type: llm\n  config:\n    provider: openai\n    model: gpt-4o\n    prompt: "Rate the response..."\n`

function insertSnippet(snippet: string) {
  content.value = content.value + '\n' + snippet
}

// Initialize
watch(currentProjectName, () => {
  selectedFile.value = null
  content.value = ''
  originalContent.value = ''
  loadTree()
}, { immediate: true })

onMounted(async () => {
  try {
    agentTypes.value = await listAgentTypes()
    graderTypes.value = await listGraderTypes()
  } catch { /* non-critical */ }
})
</script>

<template>
  <div style="height: calc(100vh - 48px); display: flex; flex-direction: column">
    <NSpace justify="space-between" align="center" style="margin-bottom: 12px; flex-shrink: 0">
      <NText tag="h1" style="margin: 0">Configurations</NText>
    </NSpace>

    <div v-if="currentProjectName" style="flex: 1; display: flex; gap: 12px; min-height: 0">
      <!-- File Tree -->
      <div style="width: 240px; flex-shrink: 0; border: 1px solid #e0e0e6; border-radius: 4px; overflow: hidden">
        <FileTree
          :tree="tree"
          :selected-file="selectedFile"
          @select-file="handleSelectFile"
          @create-file="handleCreateFile"
          @create-dir="handleCreateDir"
        />
      </div>

      <!-- Editor Panel -->
      <div v-if="selectedFile" style="flex: 1; display: flex; flex-direction: column; min-width: 0">
        <!-- Top bar -->
        <NSpace justify="space-between" align="center" style="margin-bottom: 8px; flex-shrink: 0">
          <NSpace align="center">
            <NText strong>{{ selectedFile }}</NText>
            <NTag v-if="validation.valid" type="success" size="small">Valid</NTag>
            <NTag v-else type="error" size="small">{{ validation.errors.length }} error(s)</NTag>
          </NSpace>
          <NButton type="primary" size="small" :loading="saving" :disabled="!isDirty" @click="handleSave">
            Save
          </NButton>
        </NSpace>

        <!-- Validation errors -->
        <NAlert
          v-if="!validation.valid && validation.errors.length > 0"
          type="error"
          title="Validation Errors"
          style="margin-bottom: 8px; flex-shrink: 0"
        >
          <ul style="margin: 0; padding-left: 20px">
            <li v-for="(err, i) in validation.errors" :key="i">{{ err }}</li>
          </ul>
        </NAlert>

        <!-- Editor + Assist -->
        <div style="flex: 1; display: flex; gap: 12px; min-height: 0">
          <div style="flex: 1; min-width: 0">
            <YamlEditor v-model="content" />
          </div>
          <div style="width: 220px; flex-shrink: 0; overflow-y: auto">
            <NSpace vertical :size="8">
              <NCard title="Quick Insert" size="small">
                <NSpace vertical :size="6">
                  <NButton block size="tiny" @click="insertSnippet(agentTemplate)">Agent Template</NButton>
                  <NButton block size="tiny" @click="insertSnippet(taskTemplate)">Task Template</NButton>
                  <NButton block size="tiny" @click="insertSnippet(graderTemplate)">Grader Template</NButton>
                </NSpace>
              </NCard>
              <NCard title="Agent Types" size="small">
                <NList :show-divider="false" size="small">
                  <NListItem v-for="t in agentTypes" :key="t">
                    <NText code style="font-size: 12px">{{ t }}</NText>
                  </NListItem>
                </NList>
              </NCard>
              <NCard title="Grader Types" size="small">
                <NList :show-divider="false" size="small">
                  <NListItem v-for="t in graderTypes" :key="t">
                    <NText code style="font-size: 12px">{{ t }}</NText>
                  </NListItem>
                </NList>
              </NCard>
            </NSpace>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else style="flex: 1; display: flex; align-items: center; justify-content: center">
        <NEmpty description="Select a file from the tree to edit" />
      </div>
    </div>

    <NEmpty v-else description="Select a project first" />
  </div>
</template>
