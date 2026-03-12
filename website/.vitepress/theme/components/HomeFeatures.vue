<script setup>
import { ref, onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = () => lang.value?.startsWith('zh')

const cards = ref([])

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, { threshold: 0.15 })

  document.querySelectorAll('.feature-card').forEach(el => observer.observe(el))
})

const features = [
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3"/></svg>`,
    titleEn: 'pass@k / pass^k Metrics',
    titleZh: 'pass@k / pass^k 度量',
    descEn: 'Measure capability ceiling and reliability with statistically rigorous metrics. Log-space arithmetic prevents overflow for large sample sizes.',
    descZh: '使用对数空间算法精确计算能力上限和可靠性，为 Agent 生产部署提供统计学依据。',
    color: '#5b6cf0'
  },
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>`,
    titleEn: '8 Built-in Graders',
    titleZh: '8 种内置评分器',
    descEn: 'exact_match, contains, regex, json_match, command, llm, pairwise, constraint — from simple string checks to LLM-as-judge.',
    descZh: 'exact_match、contains、regex、json_match、command、llm、pairwise、constraint — 从简单字符串匹配到 LLM 评判。',
    color: '#22d3ee'
  },
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/></svg>`,
    titleEn: '4 Agent Adapters',
    titleZh: '4 种 Agent 适配器',
    descEn: 'Native OpenAI, Anthropic, HTTP, and Command adapters. Registry pattern makes adding custom adapters a single-file change.',
    descZh: '原生支持 OpenAI、Anthropic、HTTP 和 Command，注册模式让自定义适配器只需一个文件。',
    color: '#a78bfa'
  },
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
    titleEn: 'Token / Cost Tracking',
    titleZh: 'Token / 成本追踪',
    descEn: 'Automatic token extraction with cost estimation. P50/P90/P99 latency percentiles for SLA assessment.',
    descZh: '自动提取 Token 用量并估算成本，P50/P90/P99 延迟百分位用于 SLA 评估。',
    color: '#f59e0b'
  },
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9"/></svg>`,
    titleEn: 'Cache & Checkpoints',
    titleZh: '缓存与断点续评',
    descEn: 'File-based response caching avoids redundant API calls. Checkpoint resume picks up interrupted evaluations seamlessly.',
    descZh: '基于文件的响应缓存避免重复 API 调用，断点续评从中断处无缝恢复。',
    color: '#10b981'
  },
  {
    icon: `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20V10"/><path d="M18 20V4"/><path d="M6 20v-4"/></svg>`,
    titleEn: 'CI/CD Integration',
    titleZh: 'CI/CD 集成',
    descEn: 'Use --fail-under to gate merges on pass rate. JSON output and summary files for automated pipeline processing.',
    descZh: '通过 --fail-under 阈值门禁控制合并，JSON 输出支持流水线自动化处理。',
    color: '#ef4444'
  }
]
</script>

<template>
  <section class="features-section home-section">
    <div class="section-header animate-on-scroll">
      <h2 class="section-title">
        {{ isZh() ? '核心能力' : 'Core Features' }}
      </h2>
      <p class="section-desc">
        {{ isZh()
          ? '专为 AI Agent 评估设计的完整工具链'
          : 'A complete toolkit designed for AI agent evaluation'
        }}
      </p>
    </div>
    <div class="features-grid">
      <div
        v-for="(f, i) in features"
        :key="i"
        class="feature-card animate-on-scroll"
        :style="{ '--delay': i * 100 + 'ms', '--accent': f.color }"
      >
        <div class="card-icon" v-html="f.icon"></div>
        <h3 class="card-title">{{ isZh() ? f.titleZh : f.titleEn }}</h3>
        <p class="card-desc">{{ isZh() ? f.descZh : f.descEn }}</p>
        <div class="card-shine"></div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.features-section {
  padding: var(--ae-section-gap) 24px;
}

.section-header {
  text-align: center;
  margin-bottom: 48px;
}

.section-title {
  font-size: 36px;
  font-weight: 700;
  letter-spacing: -0.5px;
  margin-bottom: 12px;
}

.section-desc {
  font-size: 17px;
  color: var(--vp-c-text-2);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.feature-card {
  position: relative;
  padding: 28px;
  border-radius: 16px;
  background: var(--ae-card-bg);
  border: 1px solid var(--ae-card-border);
  backdrop-filter: blur(12px);
  overflow: hidden;
  transition: transform 0.3s ease, box-shadow 0.3s ease, border-color 0.3s ease;
  animation-delay: var(--delay);
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.08);
  border-color: var(--accent);
}

.feature-card:hover .card-shine {
  opacity: 1;
}

.card-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
  opacity: 0;
  transition: opacity 0.3s;
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  color: var(--accent);
  margin-bottom: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.card-desc {
  font-size: 14px;
  line-height: 1.6;
  color: var(--vp-c-text-2);
}

@media (max-width: 960px) {
  .features-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 640px) {
  .features-grid { grid-template-columns: 1fr; }
  .section-title { font-size: 28px; }
}
</style>
