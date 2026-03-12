<script setup>
import { ref, onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = () => lang.value?.startsWith('zh')

const counters = ref([
  { value: 0, target: 8, suffix: '', labelEn: 'Built-in Graders', labelZh: '内置评分器' },
  { value: 0, target: 4, suffix: '', labelEn: 'Agent Adapters', labelZh: 'Agent 适配器' },
  { value: 0, target: 4, suffix: '', labelEn: 'Report Formats', labelZh: '报告格式' },
  { value: 0, target: 0, suffix: '', labelEn: 'CGO Dependencies', labelZh: 'CGO 依赖' }
])

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
        animateCounters()
        observer.unobserve(entry.target)
      }
    })
  }, { threshold: 0.3 })

  const el = document.querySelector('.stats-section')
  if (el) observer.observe(el)
})

function animateCounters() {
  counters.value.forEach((counter, i) => {
    const duration = 1500
    const steps = 30
    const increment = counter.target / steps
    let current = 0
    const interval = setInterval(() => {
      current += increment
      if (current >= counter.target) {
        counters.value[i].value = counter.target
        clearInterval(interval)
      } else {
        counters.value[i].value = Math.round(current)
      }
    }, duration / steps)
  })
}
</script>

<template>
  <section class="stats-section">
    <div class="stats-inner home-section">
      <div class="stats-grid">
        <div
          v-for="(stat, i) in counters"
          :key="i"
          class="stat-card"
          :style="{ '--delay': i * 100 + 'ms' }"
        >
          <div class="stat-value">
            {{ stat.value }}{{ stat.suffix }}
          </div>
          <div class="stat-label">
            {{ isZh() ? stat.labelZh : stat.labelEn }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.stats-section {
  padding: var(--ae-section-gap) 0;
  background: var(--vp-c-bg-soft);
  border-top: 1px solid var(--ae-card-border);
  border-bottom: 1px solid var(--ae-card-border);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}

.stat-card {
  text-align: center;
  padding: 32px 16px;
}

.stat-value {
  font-size: 48px;
  font-weight: 800;
  background: var(--ae-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1.1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 15px;
  color: var(--vp-c-text-2);
  font-weight: 500;
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .stat-value {
    font-size: 36px;
  }
}
</style>
