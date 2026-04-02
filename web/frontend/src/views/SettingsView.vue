<script setup lang="ts">
import { ref, watch } from 'vue'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { getProjectInfo } from '@/api/projects'
import type { ProjectInfo } from '@/api/projects'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

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
  <div class="space-y-6">
    <h1 class="text-xl font-semibold text-zinc-900 tracking-tight">Settings</h1>

    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm">Project Information</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="divide-y divide-gray-200">
          <div class="flex py-3">
            <span class="w-40 text-sm font-medium text-zinc-900">Project Name</span>
            <span class="text-sm text-muted-foreground">{{ info?.name || 'None selected' }}</span>
          </div>
          <div class="flex py-3">
            <span class="w-40 text-sm font-medium text-zinc-900">Project Path</span>
            <span class="text-sm text-muted-foreground">{{ info?.path || 'N/A' }}</span>
          </div>
          <div class="flex py-3">
            <span class="w-40 text-sm font-medium text-zinc-900">Database Path</span>
            <span class="text-sm text-muted-foreground">{{ info?.db_path || 'N/A' }}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
