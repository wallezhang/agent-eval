<script setup>
import { onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = () => lang.value?.startsWith('zh')

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, { threshold: 0.1 })
  document.querySelectorAll('.arch-section .animate-on-scroll').forEach(el => observer.observe(el))
})
</script>

<template>
  <section class="arch-section home-section">
    <div class="section-header animate-on-scroll">
      <h2 class="section-title">
        {{ isZh() ? '评估流程' : 'How It Works' }}
      </h2>
      <p class="section-desc">
        {{ isZh()
          ? '从 YAML 配置到可视化报告，完全自动化的评估流水线'
          : 'From YAML config to visual reports — a fully automated evaluation pipeline'
        }}
      </p>
    </div>

    <div class="flow-pipeline animate-on-scroll">
      <div class="flow-step" v-for="(step, i) in [
        { iconEn: 'YAML', iconZh: 'YAML', titleEn: 'Load Config', titleZh: '加载配置', descEn: 'Parse YAML, expand env vars, apply defaults', descZh: '解析 YAML，展开环境变量，应用默认值' },
        { iconEn: 'Agent', iconZh: 'Agent', titleEn: 'Create Agent', titleZh: '创建 Agent', descEn: 'Initialize the agent adapter from config', descZh: '根据配置初始化 Agent 适配器' },
        { iconEn: 'Run', iconZh: '执行', titleEn: 'Run Trials', titleZh: '运行试验', descEn: 'Concurrent execution with rate limiting', descZh: '并发执行，支持速率限制' },
        { iconEn: 'Grade', iconZh: '评分', titleEn: 'Grade Results', titleZh: '评分', descEn: 'Apply graders, compute weighted scores', descZh: '应用评分器，计算加权分数' },
        { iconEn: 'Report', iconZh: '报告', titleEn: 'Generate Reports', titleZh: '生成报告', descEn: 'Table, JSON, HTML with pass@k metrics', descZh: '表格/JSON/HTML 及 pass@k 指标' }
      ]" :key="i" :style="{ '--step-delay': i * 150 + 'ms' }">
        <div class="step-icon">{{ isZh() ? step.iconZh : step.iconEn }}</div>
        <div class="step-title">{{ isZh() ? step.titleZh : step.titleEn }}</div>
        <div class="step-desc">{{ isZh() ? step.descZh : step.descEn }}</div>
        <div v-if="i < 4" class="step-arrow">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 12h14M12 5l7 7-7 7"/>
          </svg>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.arch-section {
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

.flow-pipeline {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 0;
  flex-wrap: wrap;
}

.flow-step {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  flex: 0 0 auto;
  width: 160px;
  padding: 0 8px;
}

.step-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  font-family: var(--vp-font-family-mono);
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
  border: 2px solid rgba(91, 108, 240, 0.2);
  margin-bottom: 12px;
  transition: transform 0.3s, box-shadow 0.3s;
}

.flow-step:hover .step-icon {
  transform: scale(1.1);
  box-shadow: 0 8px 24px rgba(91, 108, 240, 0.2);
}

.step-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 4px;
}

.step-desc {
  font-size: 12px;
  color: var(--vp-c-text-3);
  line-height: 1.5;
}

.step-arrow {
  position: absolute;
  right: -16px;
  top: 18px;
  color: var(--vp-c-text-3);
  opacity: 0.4;
}

@media (max-width: 768px) {
  .flow-pipeline {
    flex-direction: column;
    align-items: center;
    gap: 24px;
  }
  .step-arrow {
    position: relative;
    right: auto;
    top: auto;
    transform: rotate(90deg);
    margin-top: 8px;
  }
}
</style>
