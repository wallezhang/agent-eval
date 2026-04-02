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

const accentColors = ['accent-left-orange', 'accent-left-green', 'accent-left-blue', 'accent-left-amber']
</script>

<template>
  <div class="grid grid-cols-4 gap-4 stagger-children">
    <Card
      v-for="(card, index) in cards"
      :key="card.label"
      class="card-shadow card-hover animate-fade-in-up"
      :class="accentColors[index % accentColors.length]"
    >
      <CardContent class="pt-5 pb-4">
        <p class="text-[11px] font-medium text-muted-foreground uppercase tracking-widest">{{ card.label }}</p>
        <p class="text-3xl font-extrabold text-zinc-900 mt-1.5 font-display tracking-tight">
          {{ card.value }}<span v-if="card.suffix" class="text-lg font-normal text-muted ml-0.5">{{ card.suffix }}</span>
        </p>
      </CardContent>
    </Card>
  </div>
</template>
