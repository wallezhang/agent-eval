# agent-eval

[![Go Reference](https://pkg.go.dev/badge/github.com/wallezhang/agent-eval.svg)](https://pkg.go.dev/github.com/wallezhang/agent-eval)
[![Go Report Card](https://goreportcard.com/badge/github.com/wallezhang/agent-eval)](https://goreportcard.com/report/github.com/wallezhang/agent-eval)
[![License](https://img.shields.io/github/license/wallezhang/agent-eval)](LICENSE)

通用智能体评测框架。通过 YAML 配置驱动，支持多种 Agent 类型和评分策略，结果持久化到 SQLite，支持表格/JSON/HTML 报告输出。

方法论参考 Anthropic [Demystifying Evals for AI Agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents)。

[English](README.md)

## 核心亮点

- **可靠性指标 (pass@k / pass^k)** — 实现 Anthropic 方法论，同时衡量 Agent 能力上限和重复试验下的可靠性
- **8 种内置评分器** — exact_match、contains、regex、json_match、command、llm、pairwise、constraint — 覆盖确定性检查、LLM 评分、A/B 对比和合规规则
- **Token/成本追踪** — 自动从 Agent 元数据中提取 Token 用量，按任务和全局估算成本
- **延迟分位数 (P50/P90/P99)** — 衡量 Agent 响应时间分布，用于 SLA 评估
- **响应缓存** — 基于文件的缓存机制，迭代评分逻辑时避免重复 API 调用
- **CI/CD 集成** — `--fail-under` 标志支持在自动化流水线中做通过率门控，同时输出机器可读的 `summary.json`
- **标签过滤** — 按标签运行任务子集（如 `--tags safety` 或 `--exclude-tags slow`）
- **失败重试** — 可配置的指数退避重试，应对 API 瞬时故障
- **断点续跑** — 通过 `--resume` 从中断处恢复评测
- **生命周期 Hooks** — 在运行/任务/试验的前后执行自定义 Shell 命令
- **4 种 Agent 适配器** — OpenAI、Anthropic、HTTP、Command — 加上扩展注册表支持自定义 Agent
- **4 种报告格式** — 表格（stdout）、JSON、HTML、Diff 对比

## 安装

```bash
go install github.com/wallezhang/agent-eval@latest
```

或从源码构建：

```bash
git clone https://github.com/wallezhang/agent-eval.git
cd agent-eval
make build
```

## 快速开始

### 1. 初始化项目

```bash
agent-eval init my-eval
cd my-eval
```

生成如下结构：

```
my-eval/
├── eval.yaml          # 评测配置
├── tasks/
│   └── sample.yaml    # 示例任务
└── results/           # 报告输出目录
```

### 2. 编辑配置

`eval.yaml`:

```yaml
name: "my-eval"
description: "Agent evaluation suite"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    temperature: 0.0
  cost_per_input_token: 0.00003     # 可选，用于成本估算
  cost_per_output_token: 0.00006

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true

execution:
  concurrency: 4
  rate_limit_rps: 10
  timeout: 60s
  max_retries: 2          # 瞬时错误重试次数（默认 0）
  retry_delay: 1s         # 初始重试延迟（默认 1s）

cache:                      # 可选，响应缓存
  enabled: true
  dir: .cache/
  ttl: 24h

task_files:
  - tasks/*.yaml

output:
  format: all       # table | json | html | all
  dir: ./results
```

`tasks/sample.yaml`:

```yaml
- id: capital-of-france
  name: "法国首都"
  tags: [geography]
  input:
    prompt: "法国的首都是哪里？只回答城市名。"
  expected:
    text: "巴黎"
```

### 3. 运行评测

```bash
agent-eval run -c eval.yaml
```

输出示例：

```
=== Evaluation Report: my-eval ===
Agent: openai | Run ID: a1b2c3d4
Duration: 3250ms

TASK      PASS  FAIL  ERR  AVG SCORE  PASS@K  PASS^K  P50ms  P90ms  P99ms
----      ----  ----  ---  ---------  ------  ------  -----  -----  -----
法国首都  3     0     0    1.000      1.000   1.000   980    1050   1080

--- Summary ---
Tasks: 1 | Trials: 3 (passed: 3, failed: 0, error: 0)
Overall Pass Rate: 100.0% | Avg Score: 1.000
Avg pass@k: 1.000 | Avg pass^k: 1.000

--- Token Usage ---
Total Input Tokens: 150 | Total Output Tokens: 18 | Total: 168
Estimated Cost: $0.0056
```

### 4. CI/CD 集成

```bash
# 通过率低于 80% 时使流水线失败
agent-eval run -c eval.yaml --fail-under 0.8

# 只运行 safety 标签的任务
agent-eval run -c eval.yaml --tags safety

# 恢复中断的运行
agent-eval run -c eval.yaml --resume <run-id>
```

报告目录中始终会输出 `summary.json` 文件，供机器消费。

## CLI 命令

| 命令 | 说明 |
|------|------|
| `agent-eval run -c <config>` | 执行评测套件 |
| `agent-eval list [--db path]` | 列出历史运行记录 |
| `agent-eval compare <runA> <runB>` | 对比两次运行结果 |
| `agent-eval init [directory]` | 初始化评测项目脚手架 |

### run

```bash
agent-eval run -c eval.yaml [flags]
```

| 标志 | 说明 | 默认值 |
|------|------|--------|
| `-c, --config` | 配置文件路径 | `eval.yaml` |
| `--db` | SQLite 数据库路径 | `<output_dir>/agent-eval.db` |
| `--verbose` | 输出详细日志 | `false` |
| `--fail-under` | 最低通过率（0.0-1.0），低于此值时退出码为 1 | `0`（禁用） |
| `--tags` | 只运行匹配标签的任务（逗号分隔） | |
| `--exclude-tags` | 排除匹配标签的任务（逗号分隔） | |
| `--no-cache` | 跳过响应缓存 | `false` |
| `--resume` | 通过运行 ID 恢复之前的运行 | |

### list

```bash
agent-eval list [--db results/agent-eval.db]
```

```
ID        SUITE        AGENT    TASKS  PASS RATE  DURATION  DATE
eefc2b36  simple-eval  command  3      33.3%      12ms      2026-02-27 23:31
```

### compare

```bash
agent-eval compare eefc2b36 b3c4d5e6
```

支持 ID 前缀匹配，输出逐任务的分数对比和回归/改进标注。

### init

```bash
agent-eval init my-project
```

在指定目录下生成 `eval.yaml` 和 `tasks/sample.yaml` 模板。

## 配置参考

### YAML 配置完整结构

```yaml
name: "suite-name"                    # 必填，套件名称
description: "..."                    # 可选，描述

agent:                                # 必填，被测 Agent 配置
  type: openai                        # Agent 类型
  config:                             # 类型相关配置
    model: gpt-4
    api_key: ${OPENAI_API_KEY}        # 支持环境变量展开
  cost_per_input_token: 0.00003       # 可选，用于成本估算
  cost_per_output_token: 0.00006      # 可选，用于成本估算

defaults:                             # 可选，全局默认值
  trials_per_task: 3                  # 每个任务重复次数（默认 1）
  pass_threshold: 0.5                 # 通过阈值（默认 0.5）
  graders:                            # 默认评分器
    - type: exact_match
      weight: 1.0                     # 权重（默认 1.0）
      config: {}

execution:                            # 可选，执行控制
  concurrency: 4                      # 并发数（默认 1）
  rate_limit_rps: 10                  # 每秒请求数限制（0=不限制）
  timeout: 60s                        # 单次试验超时
  max_retries: 2                      # 瞬时错误重试次数（默认 0）
  retry_delay: 1s                     # 初始重试延迟，每次翻倍（默认 1s）

cache:                                # 可选，响应缓存
  enabled: true                       # 启用/禁用缓存（默认 false）
  dir: .cache/                        # 缓存目录（默认 .cache/）
  ttl: 24h                            # 缓存条目 TTL（默认 24h）

hooks:                                # 可选，生命周期 Hooks（Shell 命令）
  before_run: "echo '开始评测'"
  after_run: "curl -X POST https://slack.example.com/webhook -d @-"
  before_task: ""
  after_task: ""
  before_trial: ""
  after_trial: ""

task_files:                           # 可选，外部任务文件（支持 glob）
  - tasks/*.yaml

tasks:                                # 可选，内联任务定义
  - id: task-1
    name: "任务名"
    tags: [tag1, tag2]
    trials_per_task: 5                # 可覆盖默认值
    step_limit: 10                    # 可选，期望最大步数，用于效率追踪
    input:
      prompt: "..."
      system: "..."                   # 可选，系统提示
      messages:                       # 可选，多轮对话
        - role: user
          content: "..."
    expected:
      text: "期望文本"
      fields:                         # JSON 字段匹配
        name: "value"
    graders:                          # 可覆盖默认评分器
      - type: llm
        weight: 0.5
        config:
          rubric: "评分标准"

output:                               # 可选，报告配置
  format: all                         # table | json | html | all
  dir: ./results                      # 输出目录
```

### Agent 类型

| 类型 | 说明 | 必填配置 |
|------|------|----------|
| `openai` | OpenAI Chat Completions API | `api_key` |
| `anthropic` | Anthropic Messages API | `api_key` |
| `http` | 通用 HTTP API | `url` |
| `command` | 外部命令（通过 stdin/stdout 交互） | `command` |

**openai**

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1   # 可选
    model: gpt-4                           # 可选，默认 gpt-4
    temperature: 0.0                       # 可选，默认 0.0
  cost_per_input_token: 0.00003            # 可选，用于成本报告
  cost_per_output_token: 0.00006
```

**anthropic**

```yaml
agent:
  type: anthropic
  config:
    api_key: ${ANTHROPIC_API_KEY}
    base_url: https://api.anthropic.com    # 可选
    model: claude-sonnet-4-20250514       # 可选
    temperature: 0.0
    max_tokens: 4096
  cost_per_input_token: 0.000003
  cost_per_output_token: 0.000015
```

**http**

```yaml
agent:
  type: http
  config:
    url: http://localhost:8080/api/chat
    method: POST                           # 可选，默认 POST
    headers:
      Authorization: "Bearer ${TOKEN}"
    response_path: text                    # 可选，JSON 响应中提取字段
```

**command**

```yaml
agent:
  type: command
  config:
    command: python
    args: ["-m", "my_agent"]
    working_dir: /path/to/project      # 可选，命令执行的工作目录
    timeout: 120s
    env:
      MODEL_PATH: /path/to/model
```

任务的 `prompt` 默认通过 stdin 传入，stdout 作为 Agent 输出。如果 `args` 中包含 `{{.Prompt}}`，则 prompt 会替换到参数中，不再通过 stdin 传入：

```yaml
agent:
  type: command
  config:
    command: echo
    args: ["{{.Prompt}}"]
```

### 评分器类型

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| `exact_match` | 精确字符串匹配 | 有明确答案的问答 |
| `contains` | 包含关键字检查 | 半结构化输出 |
| `regex` | 正则表达式匹配 | 格式化输出验证 |
| `json_match` | JSON 字段值匹配 | API 响应验证 |
| `command` | 外部命令评分 | 编码 Agent（单元测试） |
| `llm` | LLM 评分 + rubric | 开放式输出评估 |
| `pairwise` | A/B 成对比较 | 模型间对比 |
| `constraint` | 合规/策略检查 | 安全、PII、字数限制 |

**exact_match**

```yaml
graders:
  - type: exact_match
    config:
      ignore_case: true        # 忽略大小写
      ignore_whitespace: true  # 忽略首尾空白
```

**contains**

```yaml
graders:
  - type: contains
    config:
      ignore_case: true
      keywords: ["关键词1", "关键词2"]  # 所有关键词都需匹配才算通过
```

**regex**

```yaml
graders:
  - type: regex
    config:
      pattern: "^\\d{3}-\\d{4}$"
```

**json_match**

```yaml
graders:
  - type: json_match
    config:
      ignore_case: true
# 需在任务的 expected.fields 中定义期望字段值
```

**command**

```yaml
graders:
  - type: command
    config:
      command: python
      args: ["-m", "pytest", "tests/"]
      timeout: 60s
```

评分命令通过 stdin 接收 JSON（含 `task_id`、`agent_output`、`expected`），可以通过 stdout 返回 JSON（含 `score`、`pass`、`reason`），或直接以退出码 0/非零 表示通过/失败。

**llm**

```yaml
graders:
  - type: llm
    weight: 0.5
    config:
      provider: openai           # openai | anthropic
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com/v1  # 可选，用于 OpenAI 兼容服务
      model: gpt-4
      rubric: |
        评估标准：
        1. 回答准确性
        2. 表述简洁性
```

**pairwise**

```yaml
graders:
  - type: pairwise
    config:
      provider: openai
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com/v1  # 可选，用于 OpenAI 兼容服务
      criteria: "哪个回答更准确、更完整？"
      reference: "参考答案文本"     # 可选，默认用 expected.text
```

**constraint**

约束评分器对 Agent 输出进行合规规则检查。所有检查必须通过才算通过；评分反映通过的检查占比。

```yaml
graders:
  - type: constraint
    config:
      checks:
        - name: "no_pii"
          pattern: '(?i)(ssn|credit card|social security)'
          must_not_match: true           # 匹配到则失败
        - name: "has_disclaimer"
          pattern: "(?i)disclaimer"
          must_match: true               # 未匹配到则失败
        - name: "word_limit"
          max_words: 500                 # 超过字数则失败
        - name: "min_length"
          min_words: 10                  # 字数不足则失败
```

### 加权复合评分

同一任务可配置多个评分器，通过 `weight` 设置权重：

```yaml
graders:
  - type: exact_match
    weight: 2.0                    # 权重 2.0
    config: { ignore_case: true }
  - type: llm
    weight: 0.5                    # 权重 0.5
    config: { rubric: "..." }
```

最终得分 = 加权平均。通过条件：所有评分器都通过。

### 响应缓存

启用 `cache.enabled` 后，Agent 响应会基于（Agent 类型、Agent 配置、任务输入）的哈希缓存到磁盘。适用于迭代评分逻辑的场景 — 重新运行评测时无需重复调用 Agent API。

```yaml
cache:
  enabled: true
  dir: .cache/       # 缓存目录
  ttl: 24h           # 缓存条目有效期
```

使用 `--no-cache` 可在单次运行中跳过缓存。

### 生命周期 Hooks

Hooks 允许在评测的关键节点执行 Shell 命令。每个 Hook 通过 stdin 接收 JSON 上下文对象，包含相关数据（运行信息、任务信息、试验信息，取决于 Hook 类型）。

```yaml
hooks:
  before_run: "echo '开始评测'"
  after_run: "python scripts/notify.py"
  before_trial: ""
  after_trial: "python scripts/log_trial.py"
```

### 重试配置

当 `max_retries > 0` 时，Agent 执行的瞬时错误（限流、网络超时）会自动进行指数退避重试。仅 Agent 执行错误会触发重试 — 评分失败不会重试。

```yaml
execution:
  max_retries: 3        # 每个试验最大重试次数
  retry_delay: 1s       # 初始延迟，每次翻倍（1s → 2s → 4s）
```

## 核心指标

### pass@k

至少 1 次通过的概率（乐观指标，衡量 Agent 的能力上限）：

```
pass@k = 1 - C(n-c, k) / C(n, k)
```

### pass^k

全部 k 次都通过的概率（严格指标，衡量 Agent 的可靠性）：

```
pass^k = C(c, k) / C(n, k)
```

其中 `n` = 总试验数，`c` = 通过数，`k` = 采样数。

### 延迟分位数

每个任务结果包含基于 Agent 执行时间（不含评分时间）计算的 P50、P90、P99 延迟分位数。用于评估 Agent 是否满足延迟 SLA。

### Token 用量与成本

当可用时（OpenAI 和 Anthropic Agent 会自动上报），Token 数（输入/输出）会自动从 Agent 元数据中提取。配置 `cost_per_input_token` 和 `cost_per_output_token` 后，会按任务和全局计算估算成本。

### 步数统计

Agent 步数从元数据或 Transcript 中提取。结合可选的 `step_limit` 任务字段，可衡量 Agent 效率 — Agent 实际步数与预期的对比。

## 报告格式

### 表格（stdout）

终端输出，包含逐任务的 pass/fail/error 计数、平均分、pass@k、pass^k、P50/P90/P99 延迟，以及 Token 用量汇总。

### JSON

完整的结构化数据，包含所有试验详情、评分结果和元数据，输出到 `results/<suite>-<id>.json`。

### HTML

带样式的可视化报告，包含汇总卡片（通过率、平均分、Token 用量、成本）、带延迟和用量列的任务结果表、带步数的试验详情表，输出到 `results/<suite>-<id>.html`。

### Summary JSON

机器可读的 `summary.json` 始终写入输出目录，包含运行摘要，供 CI/CD 集成使用。

## 项目结构

```
agent-eval/
├── main.go
├── go.mod
├── Makefile
├── cmd/                          # CLI 命令
│   ├── root.go
│   ├── run.go
│   ├── list.go
│   ├── compare.go
│   └── init.go
├── pkg/
│   ├── model/                    # 领域模型 + 指标计算
│   ├── config/                   # YAML 配置加载与校验
│   ├── agent/                    # Agent 接口与适配器
│   ├── grader/                   # 评分器接口与实现
│   ├── engine/                   # 评测执行引擎（并发调度）
│   ├── cache/                    # 响应缓存
│   ├── storage/                  # 结果持久化（SQLite / 内存）
│   ├── report/                   # 报告生成（表格 / JSON / HTML / Diff）
│   └── llm/                      # LLM 客户端（供评分器使用）
├── examples/
│   ├── simple/                   # 命令 Agent 示例
│   ├── multi-grader/             # 多评分器示例
│   └── coding-agent/             # 编码 Agent 示例
└── templates/
    └── report.html.tmpl
```

## 扩展

Agent 和 Grader 均通过工厂注册表扩展。在 `init()` 中注册即可：

```go
// 注册自定义 Agent
agent.Register("my-agent", func(config map[string]any) (agent.Agent, error) {
    return &MyAgent{config: config}, nil
})

// 注册自定义 Grader
grader.Register("my-grader", func(config map[string]any) (grader.Grader, error) {
    return &MyGrader{config: config}, nil
})
```

Agent 接口：

```go
type Agent interface {
    Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error)
    Close() error
}
```

Grader 接口：

```go
type Grader interface {
    Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error)
    Type() string
}
```

## 开发

```bash
make build        # 构建二进制
make test         # 运行测试
make vet          # 静态分析
make lint         # golangci-lint（需安装）
make clean        # 清理构建产物
```

运行示例：

```bash
make run
# 等同于: go run . run -c examples/simple/eval.yaml
```
