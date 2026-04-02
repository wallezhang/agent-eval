<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { LayoutDashboard, Settings, Play, FileText } from 'lucide-vue-next'
import ProjectSwitcher from './ProjectSwitcher.vue'

const router = useRouter()
const route = useRoute()

const navItems = [
  { label: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
  { label: 'Configurations', path: '/configs', icon: FileText },
  { label: 'Runs', path: '/runs', icon: Play },
]

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<template>
  <div class="flex h-screen">
    <!-- Sidebar -->
    <aside class="w-[220px] flex-shrink-0 bg-white border-r border-gray-200 flex flex-col">
      <!-- Logo -->
      <div class="px-5 py-4">
        <span class="text-base font-semibold tracking-tight text-zinc-900">Agent Eval</span>
      </div>

      <!-- Project Switcher -->
      <ProjectSwitcher />

      <!-- Navigation -->
      <nav class="flex-1 px-3 py-2 space-y-1">
        <button
          v-for="item in navItems"
          :key="item.path"
          class="flex items-center gap-3 w-full rounded-lg px-3 py-2 text-sm transition-colors"
          :class="isActive(item.path)
            ? 'bg-primary-light text-primary font-medium'
            : 'text-muted hover:bg-zinc-100'"
          @click="router.push(item.path)"
        >
          <component :is="item.icon" class="h-4 w-4" />
          {{ item.label }}
        </button>
      </nav>

      <!-- Bottom: Settings -->
      <div class="border-t border-gray-200 px-3 py-2">
        <button
          class="flex items-center gap-3 w-full rounded-lg px-3 py-2 text-sm transition-colors"
          :class="isActive('/settings')
            ? 'bg-primary-light text-primary font-medium'
            : 'text-muted hover:bg-zinc-100'"
          @click="router.push('/settings')"
        >
          <Settings class="h-4 w-4" />
          Settings
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 bg-zinc-100 overflow-y-auto p-6">
      <router-view />
    </main>
  </div>
</template>
