<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  NSelect,
  NText,
  NButton,
  NSpace,
  NModal,
  NForm,
  NFormItem,
  NInput,
  useMessage,
} from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const { projects, currentProjectName } = storeToRefs(projectStore)
const message = useMessage()

const showAddModal = ref(false)
const newName = ref('')
const newPath = ref('')
const adding = ref(false)
const nameManuallyEdited = ref(false)

onMounted(async () => {
  try {
    await projectStore.fetchProjects()
  } catch (e) {
    message.error('Failed to load projects')
  }
})

const projectOptions = computed(() =>
  projects.value.map((p) => ({ label: p.name, value: p.name })),
)

function handleSelect(value: string) {
  projectStore.selectProject(value)
}

function handlePathInput(value: string) {
  newPath.value = value
  if (!nameManuallyEdited.value) {
    // Extract last folder name from path, ignoring trailing slashes
    const trimmed = value.replace(/\/+$/, '')
    const lastSegment = trimmed.split('/').pop() || ''
    newName.value = lastSegment
  }
}

function handleNameInput(value: string) {
  newName.value = value
  nameManuallyEdited.value = value !== ''
}

async function handleAdd() {
  if (!newName.value.trim() || !newPath.value.trim()) {
    message.warning('Name and path are required')
    return
  }
  adding.value = true
  try {
    await projectStore.add(newName.value.trim(), newPath.value.trim())
    message.success(`Project "${newName.value}" added`)
    showAddModal.value = false
    newName.value = ''
    newPath.value = ''
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : 'Failed to add project')
  } finally {
    adding.value = false
  }
}

function openAddModal() {
  newName.value = ''
  newPath.value = ''
  nameManuallyEdited.value = false
  showAddModal.value = true
}
</script>

<template>
  <div style="padding: 8px 16px">
    <NSpace v-if="projects.length > 0" vertical :size="4">
      <NSelect
        :value="currentProjectName"
        :options="projectOptions"
        placeholder="Select project"
        size="small"
        @update:value="handleSelect"
      />
      <NButton size="tiny" quaternary block @click="openAddModal">
        + Add Project
      </NButton>
    </NSpace>
    <NSpace v-else vertical :size="8" align="center">
      <NText depth="3" style="font-size: 12px">No projects</NText>
      <NButton size="small" type="primary" @click="openAddModal">
        Add Project
      </NButton>
    </NSpace>

    <NModal
      v-model:show="showAddModal"
      title="Add Project"
      preset="dialog"
      positive-text="Add"
      :loading="adding"
      @positive-click="handleAdd"
    >
      <NForm>
        <NFormItem label="Project Path">
          <NInput
            :value="newPath"
            placeholder="/path/to/agent-eval/project"
            @update:value="handlePathInput"
          />
        </NFormItem>
        <NFormItem label="Project Name">
          <NInput
            :value="newName"
            placeholder="auto-filled from path"
            @update:value="handleNameInput"
          />
        </NFormItem>
      </NForm>
      <NText depth="3" style="font-size: 12px">
        Point to an existing directory created by <NText code>agent-eval init</NText>
      </NText>
    </NModal>
  </div>
</template>
