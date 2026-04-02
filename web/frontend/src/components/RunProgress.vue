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
    <div class="h-2 bg-zinc-200 rounded-full overflow-hidden">
      <div
        class="h-full bg-primary rounded-full transition-all duration-300"
        :style="{ width: percentage + '%' }"
      />
    </div>
    <div class="flex items-center gap-3">
      <span class="text-sm text-muted-foreground">{{ completed }} / {{ total }} trials</span>
      <Badge v-if="passCount > 0" class="bg-success-light text-success border-0">{{ passCount }} passed</Badge>
      <Badge v-if="failCount > 0" class="bg-warning-light text-warning border-0">{{ failCount }} failed</Badge>
      <Badge v-if="errorCount > 0" class="bg-error-light text-error border-0">{{ errorCount }} errors</Badge>
    </div>
  </div>
</template>
