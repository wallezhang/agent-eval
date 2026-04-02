<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { ChevronRight, Folder, File, Plus } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog'
import type { FileNode } from '@/types'

const props = defineProps<{
  tree: FileNode[]
  selectedFile: string | null
}>()

const emit = defineEmits<{
  selectFile: [path: string]
  createFile: [dir: string, filename: string]
  createDir: [path: string]
  deleteFile: [path: string]
}>()

const showNewFileModal = ref(false)
const showNewDirModal = ref(false)
const newFileName = ref('')
const newDirName = ref('')
const activeDir = ref('')
const expandedDirs = ref<Set<string>>(new Set())

// Auto-expand all directories on initial render
function autoExpand(nodes: FileNode[]) {
  for (const node of nodes) {
    if (node.type === 'dir') {
      expandedDirs.value.add(node.path)
      if (node.children) autoExpand(node.children)
    }
  }
}
autoExpand(props.tree)

function toggleDir(path: string) {
  if (expandedDirs.value.has(path)) {
    expandedDirs.value.delete(path)
  } else {
    expandedDirs.value.add(path)
  }
  activeDir.value = path
}

function selectFile(path: string) {
  emit('selectFile', path)
}

function getContextDir(): string {
  if (props.selectedFile) {
    const parts = props.selectedFile.split('/')
    if (parts.length > 1) return parts.slice(0, -1).join('/')
  }
  return activeDir.value
}

function openNewFileModal() {
  newFileName.value = ''
  showNewFileModal.value = true
}

function openNewDirModal() {
  newDirName.value = ''
  showNewDirModal.value = true
}

function handleCreateFile() {
  const name = newFileName.value.trim()
  if (!name) {
    toast.warning('Filename is required')
    return
  }
  const filename = name.endsWith('.yaml') || name.endsWith('.yml') ? name : `${name}.yaml`
  emit('createFile', getContextDir(), filename)
  showNewFileModal.value = false
}

function handleCreateDir() {
  const name = newDirName.value.trim()
  if (!name) {
    toast.warning('Directory name is required')
    return
  }
  const dir = getContextDir()
  const fullPath = dir ? `${dir}/${name}` : name
  emit('createDir', fullPath)
  showNewDirModal.value = false
}
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto py-2">
      <template v-if="tree.length > 0">
        <div v-for="node in tree" :key="node.path">
          <!-- Directory -->
          <template v-if="node.type === 'dir'">
            <button
              class="flex items-center gap-1.5 w-full px-2 py-1 text-sm text-muted hover:bg-zinc-100 rounded transition-colors"
              @click="toggleDir(node.path)"
            >
              <ChevronRight
                class="h-3.5 w-3.5 transition-transform flex-shrink-0"
                :class="{ 'rotate-90': expandedDirs.has(node.path) }"
              />
              <Folder class="h-3.5 w-3.5 text-primary flex-shrink-0" />
              <span class="truncate">{{ node.name }}</span>
            </button>
            <div v-if="expandedDirs.has(node.path) && node.children" class="pl-4">
              <template v-for="child in node.children" :key="child.path">
                <template v-if="child.type === 'dir'">
                  <button
                    class="flex items-center gap-1.5 w-full px-2 py-1 text-sm text-muted hover:bg-zinc-100 rounded transition-colors"
                    @click="toggleDir(child.path)"
                  >
                    <ChevronRight
                      class="h-3.5 w-3.5 transition-transform flex-shrink-0"
                      :class="{ 'rotate-90': expandedDirs.has(child.path) }"
                    />
                    <Folder class="h-3.5 w-3.5 text-primary flex-shrink-0" />
                    <span class="truncate">{{ child.name }}</span>
                  </button>
                  <div v-if="expandedDirs.has(child.path) && child.children" class="pl-4">
                    <button
                      v-for="leaf in child.children"
                      :key="leaf.path"
                      class="flex items-center gap-1.5 w-full px-2 py-1 text-sm rounded transition-colors"
                      :class="leaf.path === selectedFile ? 'bg-primary-light text-primary font-medium' : 'text-muted hover:bg-zinc-100'"
                      @click="leaf.type === 'file' ? selectFile(leaf.path) : toggleDir(leaf.path)"
                    >
                      <span class="w-3.5 flex-shrink-0" />
                      <File class="h-3.5 w-3.5 flex-shrink-0" />
                      <span class="truncate">{{ leaf.name }}</span>
                    </button>
                  </div>
                </template>
                <template v-else>
                  <button
                    class="flex items-center gap-1.5 w-full px-2 py-1 text-sm rounded transition-colors"
                    :class="child.path === selectedFile ? 'bg-primary-light text-primary font-medium' : 'text-muted hover:bg-zinc-100'"
                    @click="selectFile(child.path)"
                  >
                    <span class="w-3.5 flex-shrink-0" />
                    <File class="h-3.5 w-3.5 flex-shrink-0" />
                    <span class="truncate">{{ child.name }}</span>
                  </button>
                </template>
              </template>
            </div>
          </template>
          <!-- File at root -->
          <template v-else>
            <button
              class="flex items-center gap-1.5 w-full px-2 py-1 text-sm rounded transition-colors"
              :class="node.path === selectedFile ? 'bg-primary-light text-primary font-medium' : 'text-muted hover:bg-zinc-100'"
              @click="selectFile(node.path)"
            >
              <span class="w-3.5 flex-shrink-0" />
              <File class="h-3.5 w-3.5 flex-shrink-0" />
              <span class="truncate">{{ node.name }}</span>
            </button>
          </template>
        </div>
      </template>
      <div v-else class="flex items-center justify-center py-6 text-sm text-muted-foreground">
        No config files
      </div>
    </div>

    <!-- Bottom actions -->
    <div class="border-t border-gray-200 px-2 py-2 flex gap-1">
      <button class="text-xs text-muted hover:text-zinc-900 px-2 py-1 transition-colors" @click="openNewDirModal">
        + Folder
      </button>
      <button class="text-xs text-muted hover:text-zinc-900 px-2 py-1 transition-colors" @click="openNewFileModal">
        + File
      </button>
    </div>

    <!-- New File Dialog -->
    <Dialog v-model:open="showNewFileModal">
      <DialogContent class="sm:max-w-[350px]">
        <DialogHeader>
          <DialogTitle>New Config File</DialogTitle>
        </DialogHeader>
        <div class="space-y-2 py-4">
          <label class="text-sm text-muted-foreground">Directory: {{ getContextDir() || '(root)' }}</label>
          <Input v-model="newFileName" placeholder="sample.yaml" @keyup.enter="handleCreateFile" />
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline" size="sm">Cancel</Button>
          </DialogClose>
          <Button size="sm" @click="handleCreateFile">Create</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- New Dir Dialog -->
    <Dialog v-model:open="showNewDirModal">
      <DialogContent class="sm:max-w-[350px]">
        <DialogHeader>
          <DialogTitle>New Folder</DialogTitle>
        </DialogHeader>
        <div class="space-y-2 py-4">
          <label class="text-sm text-muted-foreground">Parent: {{ getContextDir() || '(root)' }}</label>
          <Input v-model="newDirName" placeholder="tasks" @keyup.enter="handleCreateDir" />
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline" size="sm">Cancel</Button>
          </DialogClose>
          <Button size="sm" @click="handleCreateDir">Create</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
