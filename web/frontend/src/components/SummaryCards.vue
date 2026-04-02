<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'

export interface SummaryCard {
  label: string
  value: string | number
  suffix?: string
}

defineProps<{
  cards: SummaryCard[]
}>()
</script>

<template>
  <div class="grid grid-cols-4 gap-4">
    <Card v-for="card in cards" :key="card.label">
      <CardContent class="pt-6">
        <p class="text-xs text-muted-foreground uppercase tracking-wide">{{ card.label }}</p>
        <p class="text-2xl font-bold text-zinc-900 mt-1">
          {{ card.value }}<span v-if="card.suffix" class="text-lg font-normal text-muted">{{ card.suffix }}</span>
        </p>
      </CardContent>
    </Card>
  </div>
</template>
