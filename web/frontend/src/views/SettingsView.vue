<script setup lang="ts">
import { ref, watch } from 'vue'
import { NSpace, NText, NCard, NDescriptions, NDescriptionsItem } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { getProjectInfo } from '@/api/projects'
import type { ProjectInfo } from '@/api/projects'

const projectStore = useProjectStore()
const { currentProjectName } = storeToRefs(projectStore)

const info = ref<ProjectInfo | null>(null)

async function loadInfo() {
  if (!currentProjectName.value) {
    info.value = null
    return
  }
  try {
    info.value = await getProjectInfo(currentProjectName.value)
  } catch {
    info.value = null
  }
}

watch(currentProjectName, loadInfo, { immediate: true })
</script>

<template>
  <NSpace vertical :size="16">
    <NText tag="h1" style="margin: 0">Settings</NText>
    <NCard title="Project Information" size="small">
      <NDescriptions label-placement="left" :column="1" bordered>
        <NDescriptionsItem label="Project Name">{{ info?.name || 'None selected' }}</NDescriptionsItem>
        <NDescriptionsItem label="Project Path">{{ info?.path || 'N/A' }}</NDescriptionsItem>
        <NDescriptionsItem label="Database Path">{{ info?.db_path || 'N/A' }}</NDescriptionsItem>
      </NDescriptions>
    </NCard>
  </NSpace>
</template>
