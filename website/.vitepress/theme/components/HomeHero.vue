<script setup>
import { ref, onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()

const isVisible = ref(false)
const terminalLines = ref([])
const copied = ref(false)
const isPlaying = ref(true)
const animationDone = ref(false)

const installCommand = 'curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash'

const copyInstallCommand = async () => {
  try {
    await navigator.clipboard.writeText(installCommand)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {}
}

const commandEN = 'agent-eval run -c eval.yaml'
const commandZH = 'agent-eval run -c eval.yaml'
const outputLines = [
  { text: '', delay: 0 },
  { text: 'Running evaluation suite: coding-agent-eval', delay: 400, class: 'info' },
  { text: 'Agent: openai (gpt-4) | Tasks: 12 | Trials: 36', delay: 800, class: 'dim' },
  { text: '', delay: 1000 },
  { text: 'TASK                     PASS RATE  pass@3  pass^3  AVG SCORE', delay: 1200, class: 'header' },
  { text: 'math-addition            3/3        1.000   1.000   1.00', delay: 1500, class: 'pass' },
  { text: 'code-fibonacci           3/3        1.000   1.000   0.95', delay: 1800, class: 'pass' },
  { text: 'code-sorting             2/3        0.963   0.333   0.78', delay: 2100, class: 'warn' },
  { text: 'reasoning-logic          1/3        0.800   0.033   0.52', delay: 2400, class: 'fail' },
  { text: '', delay: 2700 },
  { text: 'Overall: 75.0% pass | P50: 1.2s | Tokens: 12,450', delay: 3000, class: 'summary' },
]

let timers = []

const playAnimation = () => {
  timers.forEach(clearTimeout)
  timers = []
  terminalLines.value = []
  isPlaying.value = true
  animationDone.value = false

  let currentLine = 0
  const showNextLine = () => {
    if (currentLine < outputLines.length) {
      const line = outputLines[currentLine]
      const delay = currentLine === 0 ? 1500 : line.delay - (outputLines[currentLine - 1]?.delay || 0)
      const timer = setTimeout(() => {
        terminalLines.value.push(line)
        currentLine++
        showNextLine()
      }, delay)
      timers.push(timer)
    } else {
      isPlaying.value = false
      animationDone.value = true
    }
  }
  const startTimer = setTimeout(showNextLine, 500)
  timers.push(startTimer)
}

const replayAnimation = () => {
  playAnimation()
}

onMounted(() => {
  setTimeout(() => { isVisible.value = true }, 200)
  setTimeout(playAnimation, 2000)
})

const isZh = () => lang.value?.startsWith('zh')
</script>

<template>
  <section class="hero-section">
    <div class="hero-bg">
      <div class="hero-grid"></div>
      <div class="hero-glow hero-glow-1"></div>
      <div class="hero-glow hero-glow-2"></div>
    </div>

    <div class="hero-content" :class="{ visible: isVisible }">
      <h1 class="hero-title">
        <span class="title-line">AgentEval</span>
      </h1>

      <p class="hero-subtitle">
        {{ isZh()
          ? '用 YAML 定义评估，用数据驱动决策。支持 pass@k 可靠性度量、8 种评分器、4 种 Agent 适配器，让 AI Agent 的能力评估变得简单、可靠、可复现。'
          : 'Define evaluations in YAML, make decisions with data. Featuring pass@k reliability metrics, 8 graders, and 4 agent adapters for simple, reliable, and reproducible AI agent evaluation.'
        }}
      </p>

      <div class="hero-actions">
        <a :href="isZh() ? '/zh/guide/quick-start' : '/en/guide/quick-start'" class="btn-primary">
          {{ isZh() ? '快速开始' : 'Get Started' }}
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M6 3l5 5-5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
        <a class="btn-secondary" href="https://github.com/wallezhang/agent-eval">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
          GitHub
        </a>
      </div>

      <div class="hero-install">
        <code>
          <span class="install-prefix">$</span>
          <span class="install-text">curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash</span>
          <button class="install-copy" @click="copyInstallCommand" :title="copied ? (isZh() ? '已复制' : 'Copied') : (isZh() ? '复制' : 'Copy')">
            <svg v-if="!copied" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
          </button>
        </code>
      </div>
    </div>

    <div class="hero-terminal" :class="{ visible: isVisible }">
      <div class="terminal-chrome">
        <div class="terminal-dots">
          <span></span><span></span><span></span>
        </div>
        <span class="terminal-title">Terminal</span>
        <button v-if="animationDone" class="terminal-replay" @click="replayAnimation" :title="isZh() ? '重新播放' : 'Replay'">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
        </button>
      </div>
      <div class="terminal-body">
        <div class="terminal-prompt">
          <span class="prompt-symbol">$</span>
          <span class="prompt-text typing">{{ commandEN }}</span>
          <span class="prompt-cursor"></span>
        </div>
        <div
          v-for="(line, i) in terminalLines"
          :key="i"
          class="terminal-line"
          :class="line.class"
        >
          {{ line.text }}
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-section {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120px 24px 0;
}

/* Background */
.hero-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
}

.hero-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(91, 108, 240, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(91, 108, 240, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: linear-gradient(to bottom, black 50%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, black 50%, transparent 100%);
}

.hero-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  animation: float 8s ease-in-out infinite;
}

.hero-glow-1 {
  top: 10%;
  left: 15%;
  width: 400px;
  height: 400px;
  background: rgba(91, 108, 240, 0.12);
}

.hero-glow-2 {
  bottom: 10%;
  right: 15%;
  width: 350px;
  height: 350px;
  background: rgba(34, 211, 238, 0.10);
  animation-delay: -4s;
}

/* Content */
.hero-content {
  position: relative;
  z-index: 1;
  text-align: center;
  max-width: 720px;
  opacity: 0;
  transform: translateY(30px);
  transition: opacity 0.8s ease, transform 0.8s ease;
}

.hero-content.visible {
  opacity: 1;
  transform: translateY(0);
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
  border: 1px solid rgba(91, 108, 240, 0.15);
  margin-bottom: 24px;
}

.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--vp-c-brand-1);
  animation: pulse-glow 2s ease-in-out infinite;
  box-shadow: 0 0 8px rgba(91, 108, 240, 0.5);
}

.hero-title {
  font-size: 64px;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -2px;
  margin-bottom: 20px;
}

.title-line {
  background: var(--ae-gradient);
  background-size: 200% 200%;
  animation: gradient-shift 5s ease infinite;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-subtitle {
  font-size: 18px;
  line-height: 1.7;
  color: var(--vp-c-text-2);
  margin-bottom: 32px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

/* Actions */
.hero-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-bottom: 24px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 12px 28px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: var(--ae-gradient);
  background-size: 200% 200%;
  animation: gradient-shift 5s ease infinite;
  text-decoration: none;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(91, 108, 240, 0.35);
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  background: var(--ae-card-bg);
  border: 1px solid var(--ae-card-border);
  text-decoration: none;
  backdrop-filter: blur(8px);
  transition: transform 0.2s, border-color 0.2s;
}

.btn-secondary:hover {
  transform: translateY(-2px);
  border-color: var(--vp-c-brand-1);
}

/* Install command */
.hero-install {
  display: flex;
  justify-content: center;
  max-width: 100%;
  padding: 0 16px;
}

.hero-install code {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-family: var(--vp-font-family-mono);
  color: var(--vp-c-text-2);
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--ae-card-border);
  max-width: 100%;
  overflow: hidden;
}

.install-prefix {
  color: var(--vp-c-brand-1);
  font-weight: 700;
  user-select: none;
  flex-shrink: 0;
}

.install-text {
  overflow-x: auto;
  white-space: nowrap;
  scrollbar-width: thin;
  scrollbar-color: rgba(91, 108, 240, 0.3) transparent;
}

.install-text::-webkit-scrollbar {
  height: 3px;
}

.install-text::-webkit-scrollbar-track {
  background: transparent;
}

.install-text::-webkit-scrollbar-thumb {
  background: rgba(91, 108, 240, 0.3);
  border-radius: 3px;
}

.install-copy {
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

.install-copy:hover {
  color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}

/* Terminal */
.hero-terminal {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 640px;
  margin-top: 48px;
  border-radius: 12px;
  overflow: hidden;
  background: #1a1b26;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  opacity: 0;
  transform: translateY(40px);
  transition: opacity 1s ease 0.3s, transform 1s ease 0.3s;
}

.hero-terminal.visible {
  opacity: 1;
  transform: translateY(0);
}

.terminal-chrome {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.04);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.terminal-dots {
  display: flex;
  gap: 6px;
}

.terminal-dots span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.terminal-dots span:nth-child(1) { background: #ff5f56; }
.terminal-dots span:nth-child(2) { background: #ffbd2e; }
.terminal-dots span:nth-child(3) { background: #27c93f; }

.terminal-title {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.35);
  font-family: var(--vp-font-family-mono);
  flex: 1;
}

.terminal-replay {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border: none;
  background: none;
  color: rgba(255, 255, 255, 0.3);
  cursor: pointer;
  border-radius: 4px;
  transition: color 0.2s, background 0.2s;
}

.terminal-replay:hover {
  color: rgba(255, 255, 255, 0.7);
  background: rgba(255, 255, 255, 0.08);
}

.terminal-body {
  padding: 16px 20px;
  font-family: var(--vp-font-family-mono);
  font-size: 13px;
  line-height: 1.8;
  min-height: 260px;
}

.terminal-prompt {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #c0caf5;
}

.prompt-symbol {
  color: #27c93f;
  font-weight: 700;
}

.prompt-cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: var(--vp-c-brand-1);
  animation: blink-caret 1s step-end infinite;
  border-radius: 1px;
}

.terminal-line {
  color: #a9b1d6;
  white-space: pre;
  animation: fade-in 0.3s ease;
}

.terminal-line.info { color: #7aa2f7; }
.terminal-line.dim { color: #565f89; }
.terminal-line.header { color: #c0caf5; font-weight: 600; }
.terminal-line.pass { color: #9ece6a; }
.terminal-line.warn { color: #e0af68; }
.terminal-line.fail { color: #f7768e; }
.terminal-line.summary { color: #bb9af7; font-weight: 600; }

/* Responsive */
@media (max-width: 640px) {
  .hero-title {
    font-size: 40px;
    letter-spacing: -1px;
  }
  .hero-subtitle {
    font-size: 15px;
  }
  .hero-actions {
    flex-direction: column;
    align-items: center;
  }
  .hero-terminal {
    margin-top: 32px;
  }
  .terminal-body {
    font-size: 11px;
    min-height: 200px;
  }
}
</style>
