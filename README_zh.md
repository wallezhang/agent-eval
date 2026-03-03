# agent-eval

[![Go Reference](https://pkg.go.dev/badge/github.com/wallezhang/agent-eval.svg)](https://pkg.go.dev/github.com/wallezhang/agent-eval)
[![Go Report Card](https://goreportcard.com/badge/github.com/wallezhang/agent-eval)](https://goreportcard.com/report/github.com/wallezhang/agent-eval)
[![License](https://img.shields.io/github/license/wallezhang/agent-eval)](LICENSE)

通用智能体评测框架。通过 YAML 配置驱动，支持多种 Agent 类型和评分策略，结果持久化到 SQLite，支持表格/JSON/HTML 报告输出。

方法论参考 Anthropic [Demystifying Evals for AI Agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents)。

[English](README.md)

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

TASK      PASS  FAIL  ERR  AVG SCORE  PASS@K  PASS^K
----      ----  ----  ---  ---------  ------  ------
法国首都  3     0     0    1.000      1.000   1.000

--- Summary ---
Tasks: 1 | Trials: 3 (passed: 3, failed: 0, error: 0)
Overall Pass Rate: 100.0% | Avg Score: 1.000
Avg pass@k: 1.000 | Avg pass^k: 1.000
```

## CLI 命令

| 命令 | 说明 |
|------|------|
| `agent-eval run -c <config>` | 执行评测套件 |
| `agent-eval list [--db path]` | 列出历史运行记录 |
| `agent-eval compare <runA> <runB>` | 对比两次运行结果 |
| `agent-eval init [directory]` | 初始化评测项目脚手架 |

### run

```bash
agent-eval run -c eval.yaml [--db results/agent-eval.db] [--verbose]
```

- `-c, --config` — 配置文件路径（默认 `eval.yaml`）
- `--db` — SQLite 数据库路径（默认 `<output_dir>/agent-eval.db`）
- `--verbose` — 输出详细日志

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

task_files:                           # 可选，外部任务文件（支持 glob）
  - tasks/*.yaml

tasks:                                # 可选，内联任务定义
  - id: task-1
    name: "任务名"
    tags: [tag1, tag2]
    trials_per_task: 5                # 可覆盖默认值
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
      criteria: "哪个回答更准确、更完整？"
      reference: "参考答案文本"     # 可选，默认用 expected.text
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

## 核心指标

框架实现了 Anthropic 文章中定义的两个关键指标：

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

## 报告格式

### 表格（stdout）

默认输出到终端，包含逐任务的 pass/fail/error 计数、平均分、pass@k、pass^k，以及失败详情。

### JSON

完整的结构化数据，包含所有试验详情、评分结果和元数据，输出到 `results/<suite>-<id>.json`。

### HTML

带样式的可视化报告，包含汇总卡片和详细表格，输出到 `results/<suite>-<id>.html`。

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
│   ├── storage/                  # 结果持久化（SQLite / 内存）
│   ├── report/                   # 报告生成（表格 / JSON / HTML / Diff）
│   └── llm/                      # LLM 客户端（供评分器使用）
├── examples/
│   ├── simple/                   # 命令 Agent 示例
│   └── coding-agent/            # 编码 Agent 示例
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
