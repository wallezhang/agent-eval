<script setup lang="ts">
import { NCard, NGrid, NGridItem, NStatistic } from 'naive-ui'

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
  <NGrid :cols="4" :x-gap="16" :y-gap="16">
    <NGridItem v-for="card in cards" :key="card.label">
      <NCard size="small">
        <NStatistic :label="card.label" :value="card.value">
          <template v-if="card.suffix" #suffix>{{ card.suffix }}</template>
        </NStatistic>
      </NCard>
    </NGridItem>
  </NGrid>
</template>
