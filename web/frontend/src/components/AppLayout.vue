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
    <aside class="w-[220px] flex-shrink-0 bg-sidebar border-r border-gray-200/80 flex flex-col shadow-[1px_0_3px_rgba(0,0,0,0.03)]">
      <!-- Logo -->
      <div class="px-5 py-4 flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-amber-400 flex items-center justify-center shadow-sm">
          <span class="text-xs font-bold text-white font-display">AE</span>
        </div>
        <span class="text-base font-bold tracking-tight text-zinc-900 font-display">Agent Eval</span>
      </div>

      <!-- Project Switcher -->
      <ProjectSwitcher />

      <!-- Navigation -->
      <nav class="flex-1 px-3 py-2 space-y-0.5">
        <button
          v-for="(item, index) in navItems"
          :key="item.path"
          class="flex items-center gap-3 w-full rounded-lg px-3 py-2 text-sm transition-all duration-200 animate-slide-in relative"
          :style="{ animationDelay: index * 50 + 'ms' }"
          :class="isActive(item.path)
            ? 'bg-primary-light text-primary font-medium'
            : 'text-muted hover:bg-zinc-100 hover:text-zinc-700'"
          @click="router.push(item.path)"
        >
          <div
            v-if="isActive(item.path)"
            class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-4 bg-primary rounded-r-full"
          />
          <component :is="item.icon" class="h-4 w-4" />
          {{ item.label }}
        </button>
      </nav>

      <!-- Bottom: Settings -->
      <div class="border-t border-gray-200/80 px-3 py-2">
        <button
          class="flex items-center gap-3 w-full rounded-lg px-3 py-2 text-sm transition-all duration-200 relative"
          :class="isActive('/settings')
            ? 'bg-primary-light text-primary font-medium'
            : 'text-muted hover:bg-zinc-100 hover:text-zinc-700'"
          @click="router.push('/settings')"
        >
          <div
            v-if="isActive('/settings')"
            class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-4 bg-primary rounded-r-full"
          />
          <Settings class="h-4 w-4" />
          Settings
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 bg-warm bg-dot-pattern overflow-y-auto p-6 styled-scrollbar">
      <div class="animate-fade-in-up">
        <router-view />
      </div>
    </main>
  </div>
</template>
