<script setup lang="ts">
import { computed } from 'vue'
import { NSpace, NText, NCard, NDescriptions, NDescriptionsItem } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const { currentProject } = storeToRefs(projectStore)

const dbPath = computed(() =>
  currentProject.value ? `${currentProject.value.path}/agent-eval.db` : 'N/A',
)
</script>

<template>
  <NSpace vertical :size="16">
    <NText tag="h1" style="margin: 0">Settings</NText>
    <NCard title="Project Information" size="small">
      <NDescriptions label-placement="left" :column="1" bordered>
        <NDescriptionsItem label="Project Name">{{ currentProject?.name || 'None selected' }}</NDescriptionsItem>
        <NDescriptionsItem label="Project Path">{{ currentProject?.path || 'N/A' }}</NDescriptionsItem>
        <NDescriptionsItem label="Database Path">{{ dbPath }}</NDescriptionsItem>
      </NDescriptions>
    </NCard>
  </NSpace>
</template>
