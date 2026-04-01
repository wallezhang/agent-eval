import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Project } from '@/types'
import { listProjects, addProject, deleteProject } from '@/api/projects'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProjectName = ref<string | null>(null)
  const loading = ref(false)

  const currentProject = computed(() =>
    projects.value.find((p) => p.name === currentProjectName.value) ?? null,
  )

  async function fetchProjects() {
    loading.value = true
    try {
      projects.value = await listProjects()
      if (!currentProjectName.value && projects.value.length > 0) {
        currentProjectName.value = projects.value[0].name
      }
      if (
        currentProjectName.value &&
        !projects.value.find((p) => p.name === currentProjectName.value)
      ) {
        currentProjectName.value =
          projects.value.length > 0 ? projects.value[0].name : null
      }
    } finally {
      loading.value = false
    }
  }

  async function add(name: string, path: string) {
    await addProject(name, path)
    await fetchProjects()
    currentProjectName.value = name
  }

  async function remove(name: string) {
    await deleteProject(name)
    await fetchProjects()
  }

  function selectProject(name: string) {
    currentProjectName.value = name
  }

  return {
    projects,
    currentProjectName,
    currentProject,
    loading,
    fetchProjects,
    add,
    remove,
    selectProject,
  }
})
