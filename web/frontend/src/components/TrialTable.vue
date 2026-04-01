<script setup lang="ts">
import { NTag, NText, NSpace, NCollapse, NCollapseItem } from 'naive-ui'
import type { Trial } from '@/types'

defineProps<{
  trials: Trial[]
}>()
</script>

<template>
  <NCollapse>
    <NCollapseItem v-for="trial in trials" :key="trial.id || trial.index" :name="String(trial.index)">
      <template #header>
        <NSpace align="center" :size="12">
          <NText>Trial #{{ trial.index }}</NText>
          <NTag
            :type="trial.status === 'passed' ? 'success' : trial.status === 'failed' ? 'warning' : trial.status === 'error' ? 'error' : 'info'"
            size="small"
          >{{ trial.status }}</NTag>
          <NText depth="3">Score: {{ trial.score.toFixed(3) }}</NText>
          <NText depth="3">{{ ((trial.agent_duration_ms || trial.duration_ms) / 1000).toFixed(2) }}s</NText>
        </NSpace>
      </template>

      <NSpace vertical :size="8" style="padding: 8px 0">
        <template v-if="trial.grades && trial.grades.length > 0">
          <NText strong>Grades:</NText>
          <div v-for="(g, i) in trial.grades" :key="i" style="padding-left: 16px">
            <NSpace align="center" :size="8">
              <NTag :type="g.pass ? 'success' : 'error'" size="small">{{ g.grader_type }}</NTag>
              <NText>Score: {{ g.score.toFixed(3) }} (weight: {{ g.weight }})</NText>
              <NText v-if="g.reason" depth="3">— {{ g.reason }}</NText>
            </NSpace>
          </div>
        </template>

        <template v-if="trial.error">
          <NText type="error">Error: {{ trial.error }}</NText>
        </template>

        <template v-if="trial.agent_output?.text">
          <NText strong>Agent Output:</NText>
          <div style="background: #f5f5f5; padding: 8px 12px; border-radius: 4px; font-family: monospace; font-size: 13px; white-space: pre-wrap; max-height: 200px; overflow-y: auto;">
            {{ trial.agent_output.text }}
          </div>
        </template>
      </NSpace>
    </NCollapseItem>
  </NCollapse>
</template>
