<script setup lang="ts">
import { ref, watch } from 'vue'
import { useProjectStore } from '@/stores/project'
import { storeToRefs } from 'pinia'
import { getProjectInfo } from '@/api/projects'
import type { ProjectInfo } from '@/api/projects'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { FolderOpen, Database, Tag } from 'lucide-vue-next'

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
    <h1 class="text-2xl font-extrabold text-zinc-900 tracking-tight font-display">Settings</h1>

    <Card class="card-shadow">
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-display font-semibold">Project Information</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="divide-y divide-gray-100">
          <div class="flex items-center py-3.5 gap-3">
            <Tag class="h-4 w-4 text-primary flex-shrink-0" />
            <span class="w-36 text-sm font-semibold text-zinc-900 font-display">Project Name</span>
            <span class="text-sm text-muted-foreground">{{ info?.name || 'None selected' }}</span>
          </div>
          <div class="flex items-center py-3.5 gap-3">
            <FolderOpen class="h-4 w-4 text-primary flex-shrink-0" />
            <span class="w-36 text-sm font-semibold text-zinc-900 font-display">Project Path</span>
            <span class="text-sm text-muted-foreground font-mono">{{ info?.path || 'N/A' }}</span>
          </div>
          <div class="flex items-center py-3.5 gap-3">
            <Database class="h-4 w-4 text-primary flex-shrink-0" />
            <span class="w-36 text-sm font-semibold text-zinc-900 font-display">Database Path</span>
            <span class="text-sm text-muted-foreground font-mono">{{ info?.db_path || 'N/A' }}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
