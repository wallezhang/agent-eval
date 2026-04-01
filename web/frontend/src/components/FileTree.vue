<script setup lang="ts">
import { computed, ref, h } from 'vue'
import {
  NTree,
  NSpace,
  NButton,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NEmpty,
  NIcon,
  useMessage,
} from 'naive-ui'
import type { TreeOption } from 'naive-ui'
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

const message = useMessage()

const showNewFileModal = ref(false)
const showNewDirModal = ref(false)
const newFileName = ref('')
const newDirName = ref('')

// Track which directory is contextually "selected" for creating files/dirs
const activeDir = ref('')

function fileNodeToTreeOption(node: FileNode): TreeOption {
  const option: TreeOption = {
    key: node.path,
    label: node.name,
    isLeaf: node.type === 'file',
  }
  if (node.type === 'dir') {
    // Always provide children array (even empty) to prevent async-loading spinner
    const kids = node.children || []
    option.children = kids.map(fileNodeToTreeOption)
  }
  return option
}

const treeOptions = computed<TreeOption[]>(() =>
  props.tree.map(fileNodeToTreeOption),
)

const selectedKeys = computed(() =>
  props.selectedFile ? [props.selectedFile] : [],
)

function handleSelect(keys: Array<string | number>, option: Array<TreeOption | null>) {
  const opt = option[0]
  if (!opt || typeof keys[0] !== 'string') return

  if (opt.isLeaf) {
    emit('selectFile', keys[0])
  } else {
    // Clicking a directory sets it as active context for new file/folder
    activeDir.value = keys[0]
  }
}

// Find the parent directory of the selected file, or use activeDir
function getContextDir(): string {
  if (props.selectedFile) {
    const parts = props.selectedFile.split('/')
    if (parts.length > 1) {
      return parts.slice(0, -1).join('/')
    }
  }
  return activeDir.value
}

function handleExpandedUpdate(keys: Array<string | number>) {
  // Track last expanded dir for context
  if (keys.length > 0) {
    const last = keys[keys.length - 1]
    if (typeof last === 'string') {
      activeDir.value = last
    }
  }
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
    message.warning('Filename is required')
    return
  }
  const filename = name.endsWith('.yaml') || name.endsWith('.yml') ? name : `${name}.yaml`
  const dir = getContextDir()
  emit('createFile', dir, filename)
  showNewFileModal.value = false
}

function handleCreateDir() {
  const name = newDirName.value.trim()
  if (!name) {
    message.warning('Directory name is required')
    return
  }
  const dir = getContextDir()
  const fullPath = dir ? `${dir}/${name}` : name
  emit('createDir', fullPath)
  showNewDirModal.value = false
}
</script>

<template>
  <div style="display: flex; flex-direction: column; height: 100%">
    <div style="flex: 1; overflow-y: auto; padding: 8px 0">
      <NTree
        v-if="treeOptions.length > 0"
        :data="treeOptions"
        :selected-keys="selectedKeys"
        block-line
        expand-on-click
        :default-expand-all="true"
        @update:selected-keys="handleSelect"
        @update:expanded-keys="handleExpandedUpdate"
      />
      <NEmpty v-else description="No config files" style="padding: 24px 0" />
    </div>

    <div style="border-top: 1px solid #e0e0e6; padding: 8px">
      <NSpace :size="4">
        <NButton size="tiny" quaternary @click="openNewDirModal">
          + Folder
        </NButton>
        <NButton size="tiny" quaternary @click="openNewFileModal">
          + File
        </NButton>
      </NSpace>
    </div>

    <NModal
      v-model:show="showNewFileModal"
      title="New Config File"
      preset="dialog"
      positive-text="Create"
      @positive-click="handleCreateFile"
    >
      <NForm>
        <NFormItem :label="'Directory: ' + (getContextDir() || '(root)')">
          <NInput v-model:value="newFileName" placeholder="sample.yaml" />
        </NFormItem>
      </NForm>
    </NModal>

    <NModal
      v-model:show="showNewDirModal"
      title="New Folder"
      preset="dialog"
      positive-text="Create"
      @positive-click="handleCreateDir"
    >
      <NForm>
        <NFormItem :label="'Parent: ' + (getContextDir() || '(root)')">
          <NInput v-model:value="newDirName" placeholder="tasks" />
        </NFormItem>
      </NForm>
    </NModal>
  </div>
</template>
