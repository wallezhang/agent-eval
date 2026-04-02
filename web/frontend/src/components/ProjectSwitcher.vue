<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { storeToRefs } from 'pinia'
import { useProjectStore } from '@/stores/project'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog'

const projectStore = useProjectStore()
const { projects, currentProjectName } = storeToRefs(projectStore)

const showAddModal = ref(false)
const newName = ref('')
const newPath = ref('')
const adding = ref(false)
const nameManuallyEdited = ref(false)

onMounted(async () => {
  try {
    await projectStore.fetchProjects()
  } catch {
    toast.error('Failed to load projects')
  }
})

function handleSelect(value: any) {
  if (typeof value === 'string') projectStore.selectProject(value)
}

function handlePathInput(event: Event) {
  const value = (event.target as HTMLInputElement).value
  newPath.value = value
  if (!nameManuallyEdited.value) {
    const trimmed = value.replace(/\/+$/, '')
    const lastSegment = trimmed.split('/').pop() || ''
    newName.value = lastSegment
  }
}

function handleNameInput(event: Event) {
  const value = (event.target as HTMLInputElement).value
  newName.value = value
  nameManuallyEdited.value = value !== ''
}

async function handleAdd() {
  if (!newName.value.trim() || !newPath.value.trim()) {
    toast.warning('Name and path are required')
    return
  }
  adding.value = true
  try {
    await projectStore.add(newName.value.trim(), newPath.value.trim())
    toast.success(`Project "${newName.value}" added`)
    showAddModal.value = false
    newName.value = ''
    newPath.value = ''
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to add project')
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
  <div class="px-4 py-2">
    <template v-if="projects.length > 0">
      <div class="space-y-1">
        <Select :model-value="currentProjectName ?? undefined" @update:model-value="handleSelect">
          <SelectTrigger class="w-full h-8 text-sm">
            <SelectValue placeholder="Select project" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="p in projects" :key="p.name" :value="p.name">
              {{ p.name }}
            </SelectItem>
          </SelectContent>
        </Select>
        <button
          class="w-full text-xs text-muted hover:text-zinc-900 py-1 transition-colors"
          @click="openAddModal"
        >
          + Add Project
        </button>
      </div>
    </template>
    <template v-else>
      <div class="flex flex-col items-center gap-2 py-2">
        <span class="text-xs text-muted-foreground">No projects</span>
        <Button size="sm" @click="openAddModal">Add Project</Button>
      </div>
    </template>

    <Dialog v-model:open="showAddModal">
      <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Add Project</DialogTitle>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div class="space-y-2">
            <label class="text-sm font-medium">Project Path</label>
            <Input
              :value="newPath"
              placeholder="/path/to/agent-eval/project"
              @input="handlePathInput"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-medium">Project Name</label>
            <Input
              :value="newName"
              placeholder="auto-filled from path"
              @input="handleNameInput"
            />
          </div>
          <p class="text-xs text-muted-foreground">
            Point to an existing directory created by <code class="bg-zinc-100 px-1 py-0.5 rounded text-xs">agent-eval init</code>
          </p>
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button :disabled="adding" @click="handleAdd">
            {{ adding ? 'Adding...' : 'Add' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
