<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'

const props = defineProps<{
  completed: number
  total: number
  passCount: number
  failCount: number
  errorCount: number
}>()

const percentage = computed(() =>
  props.total > 0 ? Math.round((props.completed / props.total) * 100) : 0
)
</script>

<template>
  <div class="space-y-2">
    <div class="h-3 bg-zinc-200/60 rounded-full overflow-hidden">
      <div
        class="h-full rounded-full transition-all duration-500 ease-out"
        :style="{ width: percentage + '%' }"
        :class="percentage > 0
          ? 'bg-gradient-to-r from-primary to-amber-400 animate-gradient shadow-[0_0_8px_rgba(249,115,22,0.3)]'
          : 'bg-zinc-300'"
      />
    </div>
    <div class="flex items-center gap-3">
      <span class="text-sm text-muted-foreground font-display font-semibold">{{ completed }} / {{ total }} <span class="font-normal">trials</span></span>
      <Badge v-if="passCount > 0" class="bg-success-light text-success border-0 rounded-full shadow-sm">{{ passCount }} passed</Badge>
      <Badge v-if="failCount > 0" class="bg-warning-light text-warning border-0 rounded-full shadow-sm">{{ failCount }} failed</Badge>
      <Badge v-if="errorCount > 0" class="bg-error-light text-error border-0 rounded-full shadow-sm">{{ errorCount }} errors</Badge>
    </div>
  </div>
</template>
