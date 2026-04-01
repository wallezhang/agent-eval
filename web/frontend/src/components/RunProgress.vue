<script setup lang="ts">
import { NProgress, NSpace, NText, NTag } from 'naive-ui'

defineProps<{
  completed: number
  total: number
  passCount: number
  failCount: number
  errorCount: number
}>()
</script>

<template>
  <NSpace vertical :size="8">
    <NProgress
      type="line"
      :percentage="total > 0 ? Math.round((completed / total) * 100) : 0"
      :show-indicator="true"
    />
    <NSpace>
      <NText depth="3">{{ completed }} / {{ total }} trials</NText>
      <NTag v-if="passCount > 0" type="success" size="small">{{ passCount }} passed</NTag>
      <NTag v-if="failCount > 0" type="warning" size="small">{{ failCount }} failed</NTag>
      <NTag v-if="errorCount > 0" type="error" size="small">{{ errorCount }} errors</NTag>
    </NSpace>
  </NSpace>
</template>
