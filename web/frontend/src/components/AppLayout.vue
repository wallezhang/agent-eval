<script setup lang="ts">
import { h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NText,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import ProjectSwitcher from './ProjectSwitcher.vue'

const router = useRouter()
const route = useRoute()

const menuOptions: MenuOption[] = [
  { label: 'Dashboard', key: '/dashboard' },
  { label: 'Configurations', key: '/configs' },
  { label: 'Runs', key: '/runs' },
  { label: 'Settings', key: '/settings' },
]

function handleMenuUpdate(key: string) {
  router.push(key)
}
</script>

<template>
  <NLayout has-sider style="height: 100vh">
    <NLayoutSider bordered :width="220">
      <div style="padding: 16px; font-weight: bold; font-size: 16px">
        <NText>Agent Eval</NText>
      </div>
      <ProjectSwitcher />
      <NMenu
        :options="menuOptions"
        :value="route.path"
        @update:value="handleMenuUpdate"
      />
    </NLayoutSider>
    <NLayoutContent style="padding: 24px; overflow-y: auto">
      <slot />
    </NLayoutContent>
  </NLayout>
</template>
