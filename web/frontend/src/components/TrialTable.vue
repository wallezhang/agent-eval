<script setup lang="ts">
import { ref } from 'vue'
import { ChevronRight } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import type { Trial } from '@/types'

defineProps<{
  trials: Trial[]
}>()

const expandedTrials = ref<Set<number>>(new Set())

function toggleTrial(index: number) {
  if (expandedTrials.value.has(index)) {
    expandedTrials.value.delete(index)
  } else {
    expandedTrials.value.add(index)
  }
}

function statusVariant(status: string): string {
  if (status === 'passed') return 'bg-success-light text-success border-0 rounded-full shadow-sm'
  if (status === 'failed') return 'bg-warning-light text-warning border-0 rounded-full shadow-sm'
  if (status === 'error') return 'bg-error-light text-error border-0 rounded-full shadow-sm'
  return ''
}
</script>

<template>
  <div class="space-y-1">
    <div v-for="trial in trials" :key="trial.id || trial.index" class="border border-gray-200 rounded-lg overflow-hidden transition-all duration-200 hover:border-primary/30">
      <!-- Header (clickable) -->
      <button
        class="flex items-center gap-3 w-full px-4 py-2.5 text-sm hover:bg-zinc-50 transition-all duration-200 text-left group"
        @click="toggleTrial(trial.index)"
      >
        <div class="w-0 group-hover:w-0.5 h-5 bg-primary rounded-full transition-all duration-200 -ml-1 mr-0 group-hover:mr-1 opacity-0 group-hover:opacity-100" />
        <ChevronRight
          class="h-4 w-4 text-muted-foreground transition-transform flex-shrink-0"
          :class="{ 'rotate-90': expandedTrials.has(trial.index) }"
        />
        <span class="font-medium text-zinc-900 font-display">Trial #{{ trial.index }}</span>
        <Badge :class="statusVariant(trial.status)">{{ trial.status }}</Badge>
        <span class="text-muted-foreground">Score: {{ trial.score.toFixed(3) }}</span>
        <span class="text-muted-foreground">{{ ((trial.agent_duration_ms || trial.duration_ms) / 1000).toFixed(2) }}s</span>
      </button>

      <!-- Expanded content -->
      <div v-if="expandedTrials.has(trial.index)" class="border-t border-gray-200 px-4 py-3 space-y-3 border-l-2 border-l-primary/20 ml-2">
        <!-- Grades -->
        <template v-if="trial.grades && trial.grades.length > 0">
          <p class="text-sm font-medium text-zinc-900 font-display">Grades:</p>
          <div v-for="(g, i) in trial.grades" :key="i" class="flex items-center gap-2 pl-4">
            <Badge :class="g.pass ? 'bg-success-light text-success border-0 rounded-full shadow-sm' : 'bg-error-light text-error border-0 rounded-full shadow-sm'">
              {{ g.grader_type }}
            </Badge>
            <span class="text-sm">Score: {{ g.score.toFixed(3) }} (weight: {{ g.weight }})</span>
            <span v-if="g.reason" class="text-sm text-muted-foreground">— {{ g.reason }}</span>
          </div>
        </template>

        <!-- Error -->
        <p v-if="trial.error" class="text-sm text-error">Error: {{ trial.error }}</p>

        <!-- Agent Output -->
        <template v-if="trial.agent_output">
          <p class="text-sm font-medium text-zinc-900 font-display">Agent Output:</p>
          <pre
            v-if="trial.agent_output.text"
            class="bg-zinc-900 text-zinc-300 p-3 rounded-lg text-xs font-mono whitespace-pre-wrap max-h-[200px] overflow-y-auto styled-scrollbar"
          >{{ trial.agent_output.text }}</pre>
          <p v-else class="text-sm text-muted-foreground italic">(empty)</p>

          <!-- Metadata -->
          <template v-if="trial.agent_output.metadata && Object.keys(trial.agent_output.metadata).length > 0">
            <p class="text-sm font-medium text-zinc-900 font-display">Metadata:</p>
            <div class="bg-zinc-900 text-zinc-300 p-3 rounded-lg text-xs font-mono whitespace-pre-wrap max-h-[150px] overflow-y-auto styled-scrollbar">
              <div v-for="(val, key) in trial.agent_output.metadata" :key="String(key)">
                <code class="text-amber-400">{{ key }}</code>: {{ typeof val === 'object' ? JSON.stringify(val) : val }}
              </div>
            </div>
          </template>
        </template>

        <!-- Transcript -->
        <template v-if="trial.transcript?.steps?.length">
          <p class="text-sm font-medium text-zinc-900 font-display">Transcript:</p>
          <div class="bg-zinc-900 text-zinc-300 p-3 rounded-lg text-xs font-mono whitespace-pre-wrap max-h-[200px] overflow-y-auto styled-scrollbar">
            <div v-for="(step, si) in trial.transcript.steps" :key="si">
              <code class="text-amber-400">{{ step.role || step.type }}</code>: {{ step.content.length > 500 ? step.content.slice(0, 500) + '...' : step.content }}
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
