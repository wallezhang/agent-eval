<script setup>
import { ref, onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = () => lang.value?.startsWith('zh')

const copied = ref(false)
const installCommand = 'curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash'

const copyInstallCommand = async () => {
  try {
    await navigator.clipboard.writeText(installCommand)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {}
}

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, { threshold: 0.2 })
  document.querySelectorAll('.cta-section .animate-on-scroll').forEach(el => observer.observe(el))
})
</script>

<template>
  <section class="cta-section">
    <div class="cta-inner home-section animate-on-scroll">
      <div class="cta-bg-glow"></div>
      <h2 class="cta-title">
        {{ isZh() ? '开始评估你的 AI Agent' : 'Start Evaluating Your AI Agents' }}
      </h2>
      <p class="cta-desc">
        {{ isZh()
          ? '只需一个 YAML 文件，即可获得完整的评估报告、可靠性度量和成本追踪。'
          : 'Get comprehensive evaluation reports, reliability metrics, and cost tracking with just a single YAML file.'
        }}
      </p>
      <div class="cta-code">
        <code>
          <span class="code-prefix">$</span>
          <span class="code-text">curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash</span>
          <button class="code-copy" @click="copyInstallCommand" :title="copied ? (isZh() ? '已复制' : 'Copied') : (isZh() ? '复制' : 'Copy')">
            <svg v-if="!copied" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
          </button>
        </code>
      </div>
      <div class="cta-actions">
        <a :href="isZh() ? '/zh/guide/quick-start' : '/en/guide/quick-start'" class="cta-btn-primary">
          {{ isZh() ? '阅读文档' : 'Read the Docs' }}
        </a>
        <a href="https://github.com/wallezhang/agent-eval" class="cta-btn-secondary">
          GitHub
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.cta-section {
  padding: var(--ae-section-gap) 24px;
}

.cta-inner {
  position: relative;
  text-align: center;
  padding: 64px 32px;
  border-radius: 24px;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--ae-card-border);
  overflow: hidden;
}

.cta-bg-glow {
  position: absolute;
  top: -50%;
  left: 50%;
  transform: translateX(-50%);
  width: 600px;
  height: 400px;
  background: radial-gradient(ellipse, rgba(91, 108, 240, 0.08), transparent 70%);
  pointer-events: none;
}

.cta-title {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 16px;
  position: relative;
}

.cta-desc {
  font-size: 17px;
  color: var(--vp-c-text-2);
  max-width: 520px;
  margin: 0 auto 28px;
  line-height: 1.7;
  position: relative;
}

.cta-code {
  display: flex;
  justify-content: center;
  margin-bottom: 28px;
  position: relative;
  padding: 0 16px;
}

.cta-code code {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 10px;
  font-size: 14px;
  font-family: var(--vp-font-family-mono);
  color: var(--vp-c-text-1);
  background: var(--vp-c-bg);
  border: 1px solid var(--ae-card-border);
  max-width: 100%;
  overflow: hidden;
}

.code-prefix {
  color: var(--vp-c-brand-1);
  font-weight: 700;
  user-select: none;
  flex-shrink: 0;
}

.code-text {
  overflow-x: auto;
  white-space: nowrap;
  scrollbar-width: thin;
  scrollbar-color: rgba(91, 108, 240, 0.3) transparent;
}

.code-text::-webkit-scrollbar {
  height: 3px;
}

.code-text::-webkit-scrollbar-track {
  background: transparent;
}

.code-text::-webkit-scrollbar-thumb {
  background: rgba(91, 108, 240, 0.3);
  border-radius: 3px;
}

.code-copy {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border: none;
  background: none;
  color: var(--vp-c-text-3);
  cursor: pointer;
  border-radius: 4px;
  flex-shrink: 0;
  transition: color 0.2s, background 0.2s;
}

.code-copy:hover {
  color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}

.cta-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  position: relative;
}

.cta-btn-primary {
  padding: 12px 32px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: var(--ae-gradient);
  text-decoration: none;
  transition: transform 0.2s, box-shadow 0.2s;
}

.cta-btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(91, 108, 240, 0.35);
}

.cta-btn-secondary {
  padding: 12px 32px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  background: var(--vp-c-bg);
  border: 1px solid var(--ae-card-border);
  text-decoration: none;
  transition: transform 0.2s, border-color 0.2s;
}

.cta-btn-secondary:hover {
  transform: translateY(-2px);
  border-color: var(--vp-c-brand-1);
}

@media (max-width: 640px) {
  .cta-title { font-size: 24px; }
  .cta-actions { flex-direction: column; align-items: center; }
}
</style>
