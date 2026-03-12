<script setup>
import { ref, onMounted } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = () => lang.value?.startsWith('zh')
const activeTab = ref(0)

const tabs = [
  { labelEn: 'eval.yaml', labelZh: 'eval.yaml' },
  { labelEn: 'tasks/sample.yaml', labelZh: 'tasks/sample.yaml' },
  { labelEn: 'Output', labelZh: '输出结果' }
]

const codes = [
  `name: "coding-agent-eval"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: \${OPENAI_API_KEY}
    temperature: 0.0

defaults:
  trials_per_task: 3
  graders:
    - type: contains
      config:
        ignore_case: true

execution:
  concurrency: 4
  rate_limit_rps: 5
  timeout: 120s

output:
  format: all
  dir: ./results`,
  `- id: fibonacci
  name: "Fibonacci Function"
  tags: [coding, easy]
  input:
    prompt: |
      Write a Python function that returns
      the nth Fibonacci number.
  expected:
    text: "def fibonacci"
  graders:
    - type: contains
    - type: constraint
      config:
        checks:
          - pattern: "def fibonacci"
            must_match: true
          - max_words: 200`,
  `$ agent-eval run -c eval.yaml

Suite: coding-agent-eval | Agent: openai (gpt-4)
Tasks: 4 | Trials per task: 3 | Concurrency: 4

TASK              TRIALS  PASS   SCORE  pass@3  pass^3
fibonacci          3/3    3/3    0.95   1.000   1.000
sorting            3/3    2/3    0.78   0.963   0.333
binary-search      3/3    3/3    0.92   1.000   1.000
linked-list        3/3    1/3    0.55   0.800   0.033

Overall pass rate: 75.0%
Avg score: 0.80 | P50: 1.2s | P90: 2.8s
Tokens: 12,450 | Est. cost: $0.38

Reports saved to ./results/`
]

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, { threshold: 0.1 })
  document.querySelectorAll('.code-demo-section .animate-on-scroll').forEach(el => observer.observe(el))
})
</script>

<template>
  <section class="code-demo-section home-section">
    <div class="section-header animate-on-scroll">
      <h2 class="section-title">
        {{ isZh() ? '三步开始评估' : 'Evaluate in Three Steps' }}
      </h2>
      <p class="section-desc">
        {{ isZh()
          ? '定义配置 → 编写任务 → 运行评估，就是这么简单'
          : 'Define config → Write tasks → Run evaluation — it\'s that simple'
        }}
      </p>
    </div>

    <div class="demo-container animate-on-scroll">
      <div class="steps-col">
        <div
          v-for="(step, i) in [
            { numEn: '01', numZh: '01', titleEn: 'Configure', titleZh: '配置', descEn: 'Define your agent, graders, and execution settings in YAML', descZh: '在 YAML 中定义 Agent、评分器和执行参数' },
            { numEn: '02', numZh: '02', titleEn: 'Define Tasks', titleZh: '定义任务', descEn: 'Write evaluation tasks with expected outputs and custom graders', descZh: '编写评估任务，指定预期输出和自定义评分器' },
            { numEn: '03', numZh: '03', titleEn: 'Run & Analyze', titleZh: '运行分析', descEn: 'Execute evaluations and get detailed reports with reliability metrics', descZh: '执行评估，获取包含可靠性指标的详细报告' }
          ]"
          :key="i"
          class="step-item"
          :class="{ active: activeTab === i }"
          @click="activeTab = i"
        >
          <span class="step-num">{{ step.numEn }}</span>
          <div>
            <div class="step-title">{{ isZh() ? step.titleZh : step.titleEn }}</div>
            <div class="step-desc">{{ isZh() ? step.descZh : step.descEn }}</div>
          </div>
        </div>
      </div>

      <div class="code-col">
        <div class="code-window">
          <div class="code-chrome">
            <div class="code-tabs">
              <button
                v-for="(tab, i) in tabs"
                :key="i"
                class="code-tab"
                :class="{ active: activeTab === i }"
                @click="activeTab = i"
              >
                {{ isZh() ? tab.labelZh : tab.labelEn }}
              </button>
            </div>
          </div>
          <div class="code-body">
            <pre class="code-content"><code>{{ codes[activeTab] }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.code-demo-section {
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

.demo-container {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 32px;
  align-items: start;
}

.steps-col {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.step-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  border: 1px solid transparent;
}

.step-item:hover {
  background: var(--vp-c-bg-soft);
}

.step-item.active {
  background: var(--vp-c-brand-soft);
  border-color: rgba(91, 108, 240, 0.2);
}

.step-num {
  font-size: 24px;
  font-weight: 800;
  color: var(--vp-c-brand-1);
  opacity: 0.3;
  line-height: 1;
  min-width: 36px;
  transition: opacity 0.2s;
}

.step-item.active .step-num {
  opacity: 1;
}

.step-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 4px;
}

.step-desc {
  font-size: 13px;
  color: var(--vp-c-text-3);
  line-height: 1.5;
}

/* Code window */
.code-window {
  border-radius: 12px;
  overflow: hidden;
  background: #1a1b26;
  border: 1px solid rgba(255, 255, 255, 0.06);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.15);
}

.code-chrome {
  padding: 0 12px;
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.code-tabs {
  display: flex;
  gap: 0;
}

.code-tab {
  padding: 10px 16px;
  font-size: 12px;
  font-family: var(--vp-font-family-mono);
  color: rgba(255, 255, 255, 0.35);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
}

.code-tab:hover {
  color: rgba(255, 255, 255, 0.6);
}

.code-tab.active {
  color: #c0caf5;
  border-bottom-color: var(--vp-c-brand-1);
}

.code-body {
  padding: 20px;
  overflow-x: auto;
}

.code-content {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: #a9b1d6;
  font-family: var(--vp-font-family-mono);
}

@media (max-width: 768px) {
  .demo-container {
    grid-template-columns: 1fr;
  }
  .steps-col {
    flex-direction: row;
    overflow-x: auto;
  }
  .step-desc {
    display: none;
  }
}
</style>
