<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { NSelect, NText, useMessage } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const { projects, currentProjectName } = storeToRefs(projectStore)
const message = useMessage()

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
</script>

<template>
  <div style="padding: 8px 16px">
    <NSelect
      v-if="projects.length > 0"
      :value="currentProjectName"
      :options="projectOptions"
      placeholder="Select project"
      size="small"
      @update:value="handleSelect"
    />
    <NText v-else depth="3" style="font-size: 12px">No projects</NText>
  </div>
</template>
