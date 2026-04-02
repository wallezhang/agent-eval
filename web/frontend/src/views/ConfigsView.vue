<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { getConfig, updateConfig, createConfig, validateConfig } from '@/api/configs'
import { listFileTree, createDir } from '@/api/files'
import { listAgentTypes, listGraderTypes } from '@/api/meta'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import FileTree from '@/components/FileTree.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import type { FileNode, ValidationResult } from '@/types'

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)

const tree = ref<FileNode[]>([])
const selectedFile = ref<string | null>(null)

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

async function loadTree() {
  if (!currentProjectName.value) return
  try {
    tree.value = await listFileTree(currentProjectName.value)
  } catch {
    toast.error('Failed to load file tree')
  }
}

async function loadFile(path: string) {
  if (!currentProjectName.value) return
  try {
    const data = await getConfig(currentProjectName.value, path)
    isLoadingFile.value = true
    selectedFile.value = path
    content.value = data
    originalContent.value = data
    validation.value = { valid: true, errors: [] }
    await handleValidate()
  } catch {
    toast.error(`Failed to load ${path}`)
  } finally {
    isLoadingFile.value = false
  }
}

async function handleSave() {
  if (!currentProjectName.value || !selectedFile.value) return
  saving.value = true
  try {
    await updateConfig(currentProjectName.value, selectedFile.value, content.value)
    originalContent.value = content.value
    isDirty.value = false
    toast.success('Saved')
    await handleValidate()
  } catch {
    toast.error('Failed to save')
  } finally {
    saving.value = false
  }
}

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
  if (isLoadingFile.value) return
  if (selectedFile.value) {
    validateTimer = setTimeout(handleValidate, 800)
  }
})

function handleSelectFile(path: string) {
  loadFile(path)
}

async function handleCreateFile(dir: string, filename: string) {
  if (!currentProjectName.value) return
  const fullPath = dir ? `${dir}/${filename}` : filename
  const template = `# ${filename}\n`
  try {
    await createConfig(currentProjectName.value, fullPath, template)
    toast.success(`Created ${fullPath}`)
    await loadTree()
    loadFile(fullPath)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to create file')
  }
}

async function handleCreateDir(path: string) {
  if (!currentProjectName.value) return
  try {
    await createDir(currentProjectName.value, path)
    toast.success(`Created folder ${path}`)
    await loadTree()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to create folder')
  }
}

const agentTemplate = `agent:\n  type: openai\n  config:\n    model: gpt-4o\n    api_key: \${OPENAI_API_KEY}\n    temperature: 0\n`
const taskTemplate = `- id: task-name\n  name: "Task description"\n  input:\n    prompt: "prompt text"\n  expected:\n    text: "expected output"\n  graders:\n    - type: exact_match\n`
const graderTemplate = `- type: llm\n  config:\n    provider: openai\n    model: gpt-4o\n    prompt: "Rate the response..."\n`

function insertSnippet(snippet: string) {
  content.value = content.value + '\n' + snippet
}

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
  <div class="h-[calc(100vh-48px)] flex flex-col">
    <div class="flex items-center justify-between mb-3 flex-shrink-0">
      <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">Configurations</h1>
    </div>

    <div v-if="currentProjectName" class="flex-1 flex gap-3 min-h-0">
      <div class="w-[240px] flex-shrink-0 bg-white border border-gray-200 rounded-lg overflow-hidden">
        <FileTree
          :tree="tree"
          :selected-file="selectedFile"
          @select-file="handleSelectFile"
          @create-file="handleCreateFile"
          @create-dir="handleCreateDir"
        />
      </div>

      <div v-if="selectedFile" class="flex-1 flex flex-col min-w-0">
        <div class="flex items-center justify-between mb-2 flex-shrink-0">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-zinc-900">{{ selectedFile }}</span>
            <Badge v-if="validation.valid" class="bg-success-light text-success border-0">Valid</Badge>
            <Badge v-else class="bg-error-light text-error border-0">{{ validation.errors.length }} error(s)</Badge>
          </div>
          <Button size="sm" :disabled="!isDirty || saving" @click="handleSave">
            {{ saving ? 'Saving...' : 'Save' }}
          </Button>
        </div>

        <div
          v-if="!validation.valid && validation.errors.length > 0"
          class="mb-2 p-3 bg-error-light border border-red-200 rounded-lg text-sm text-error flex-shrink-0"
        >
          <p class="font-medium mb-1">Validation Errors</p>
          <ul class="list-disc pl-5 space-y-0.5">
            <li v-for="(err, i) in validation.errors" :key="i">{{ err }}</li>
          </ul>
        </div>

        <div class="flex-1 flex gap-3 min-h-0">
          <div class="flex-1 min-w-0">
            <YamlEditor v-model="content" />
          </div>
          <div class="w-[220px] flex-shrink-0 overflow-y-auto space-y-2">
            <Card>
              <CardHeader class="pb-2 pt-4 px-4">
                <CardTitle class="text-sm">Quick Insert</CardTitle>
              </CardHeader>
              <CardContent class="px-4 pb-4 space-y-1.5">
                <Button variant="outline" size="sm" class="w-full justify-start text-xs" @click="insertSnippet(agentTemplate)">Agent Template</Button>
                <Button variant="outline" size="sm" class="w-full justify-start text-xs" @click="insertSnippet(taskTemplate)">Task Template</Button>
                <Button variant="outline" size="sm" class="w-full justify-start text-xs" @click="insertSnippet(graderTemplate)">Grader Template</Button>
              </CardContent>
            </Card>
            <Card>
              <CardHeader class="pb-2 pt-4 px-4">
                <CardTitle class="text-sm">Agent Types</CardTitle>
              </CardHeader>
              <CardContent class="px-4 pb-4">
                <div class="space-y-1">
                  <code v-for="t in agentTypes" :key="t" class="block text-xs bg-zinc-100 px-2 py-0.5 rounded">{{ t }}</code>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader class="pb-2 pt-4 px-4">
                <CardTitle class="text-sm">Grader Types</CardTitle>
              </CardHeader>
              <CardContent class="px-4 pb-4">
                <div class="space-y-1">
                  <code v-for="t in graderTypes" :key="t" class="block text-xs bg-zinc-100 px-2 py-0.5 rounded">{{ t }}</code>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <div v-else class="flex-1 flex items-center justify-center text-sm text-muted-foreground">
        Select a file from the tree to edit
      </div>
    </div>

    <div v-else class="flex-1 flex items-center justify-center text-sm text-muted-foreground">
      Select a project first
    </div>
  </div>
</template>
