# 配置参考

AgentEval 的完整 YAML 配置参考。

## 完整配置模式

```yaml
# 套件名称（必填）
name: "my-eval"
# 描述（可选）
description: "评估套件描述"

# Agent 配置（必填）
agent:
  type: openai | anthropic | http | command
  config:
    # 类型特定配置（见下方 Agent 类型）

# 默认设置，应用于所有任务（可选）
defaults:
  trials_per_task: 1          # 每个任务的试验次数
  pass_threshold: 0.5         # 通过的最低分数
  graders:                    # 默认评分器
    - type: exact_match
      weight: 1.0
      config: {}

# 执行设置（可选）
execution:
  concurrency: 1              # 最大并发试验数
  rate_limit_rps: 0           # 每秒请求数（0 = 不限制）
  timeout: "60s"              # 单次试验超时
  retry:
    max_retries: 0            # 最大重试次数（仅针对 Agent 执行错误）
    initial_delay: "1s"       # 初始退避延迟
    max_delay: "30s"          # 最大退避延迟

# 任务文件（至少需要一个，除非使用内联任务）
task_files:
  - tasks/*.yaml

# 内联任务（task_files 的替代方式）
tasks:
  - id: task-1
    name: "任务名称"
    input:
      prompt: "你的提示词"

# 输出配置（可选）
output:
  format: table | json | html | all   # 默认：table
  dir: ./results                       # 输出目录

# 响应缓存（可选）
cache:
  enabled: false
  dir: .cache/agent-eval

# 生命周期钩子（可选）
hooks:
  before_run: ""              # 评估前执行的 shell 命令
  after_run: ""               # 评估后执行的 shell 命令
```

## Agent 类型

### openai

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}    # 必填
    model: gpt-4                   # 默认：gpt-4
    base_url: https://api.openai.com/v1  # 可选，支持兼容 API
    temperature: 0.0               # 默认：0.0
```

兼容任何 OpenAI 兼容 API（如 vLLM、Ollama、Azure OpenAI），只需设置 `base_url`。

### anthropic

```yaml
agent:
  type: anthropic
  config:
    api_key: ${ANTHROPIC_API_KEY}  # 必填
    model: claude-sonnet-4-20250514  # 默认
    base_url: https://api.anthropic.com  # 可选
    temperature: 0.0               # 可选
    max_tokens: 4096               # 默认：4096
```

### http

```yaml
agent:
  type: http
  config:
    url: https://my-api.example.com/evaluate  # 必填
    method: POST                              # 默认：POST
    headers:                                  # 可选
      Authorization: "Bearer ${API_TOKEN}"
    body_template: ""                         # 可选自定义请求体模板
    response_path: "text"                     # 响应 JSON 路径，默认："text"
```

HTTP Agent 发送包含 `prompt`、`system`、`messages` 和 `params` 字段的 JSON 请求体。响应文本从 `response_path` 指定的 JSON 路径提取，如提取失败则使用原始响应文本。

### command

```yaml
agent:
  type: command
  config:
    command: python                  # 必填
    args: ["eval_script.py"]         # 可选
    working_dir: ./scripts           # 可选
    timeout: "60s"                   # 默认：60s
    env:                             # 可选环境变量
      MY_VAR: "value"
```

默认通过 stdin 传递提示词。如果任何参数包含 <code v-pre>{{.Prompt}}</code>，则将提示词替换到参数中。

## 评分器类型

### exact_match

```yaml
- type: exact_match
  weight: 1.0
  config:
    ignore_case: false         # 忽略大小写
    ignore_whitespace: false   # 忽略首尾空白
```

将 Agent 输出文本与 `expected.text` 进行比较。匹配得 1.0 分，否则 0.0 分。

### contains

```yaml
- type: contains
  weight: 1.0
  config:
    ignore_case: false
    keywords: ["关键词1", "关键词2"]  # 额外必须包含的关键词
```

检查 Agent 输出是否包含 `expected.text` 和所有 `keywords`。分数 = 匹配数 / 总数。

### regex

```yaml
- type: regex
  weight: 1.0
  config:
    pattern: "\\d{3}-\\d{4}"    # 必填，Go 正则语法
```

对 Agent 输出测试正则表达式。匹配得 1.0 分，否则 0.0 分。

### json_match

```yaml
- type: json_match
  weight: 1.0
  config:
    ignore_case: false
```

将 Agent 输出解析为 JSON，与 `expected.fields` 进行对比。分数 = 匹配字段数 / 总字段数。

任务定义示例：

```yaml
- id: json-test
  input:
    prompt: "以 JSON 格式返回用户信息"
  expected:
    fields:
      name: "Alice"
      age: 30
  graders:
    - type: json_match
```

### command

```yaml
- type: command
  weight: 1.0
  config:
    command: python               # 必填
    args: ["grade_script.py"]
    timeout: "60s"                # 默认：60s
```

通过 stdin 接收 JSON：`{"task_id": "...", "agent_output": "...", "expected": {...}}`。退出码 0 = 通过，非零 = 失败。可通过 stdout 返回 JSON：`{"score": 0.8, "pass": true, "reason": "..."}`。

### llm

```yaml
- type: llm
  weight: 1.0
  config:
    provider: openai              # 默认：openai
    api_key: ${OPENAI_API_KEY}
    base_url: ""                  # 可选，支持兼容 API
    model: gpt-4
    rubric: |                     # 必填
      请根据以下标准评估回答：
      1. 正确性
      2. 完整性
```

使用 LLM 根据评分标准评估 Agent 输出。LLM 响应需包含：`SCORE: <0.0-1.0>`、`PASS: <true/false>`、`REASON: <text>`。

### pairwise

```yaml
- type: pairwise
  weight: 1.0
  config:
    provider: openai
    api_key: ${OPENAI_API_KEY}
    base_url: ""                  # 可选
    model: gpt-4
    criteria: ""                  # 可选，默认比较准确性/完整性/有用性
    reference: ""                 # 可选，未设置时使用 expected.text
```

使用 LLM 将 Agent 输出与参考答案进行比较。响应格式：`VERDICT: <A_BETTER|B_BETTER|TIE>`、`REASON: <text>`。分数：B_BETTER=1.0、TIE=0.5、A_BETTER=0.0。

### constraint

```yaml
- type: constraint
  weight: 1.0
  config:
    checks:
      - pattern: "\\d+"
        must_match: true
      - pattern: "error"
        must_not_match: true
      - max_words: 100
      - min_words: 10
```

对 Agent 输出应用结构约束。所有检查必须通过。分数 = 通过数 / 总检查数。

## 任务文件格式

任务文件是 YAML 数组：

```yaml
- id: task-1                       # 必填，唯一
  name: "任务名称"                  # 可选
  tags: [math, easy]               # 可选，用于过滤
  input:
    prompt: "你的提示词"            # 必填（或使用 messages）
    system: "系统提示词"            # 可选
    messages:                      # prompt 的替代方式
      - role: user
        content: "你好"
      - role: assistant
        content: "你好！"
    params:                        # 可选键值参数
      language: "python"
  expected:                        # 可选
    text: "预期输出"
    fields:                        # 用于 json_match
      key: "value"
  graders:                         # 可选，覆盖默认评分器
    - type: exact_match
  trials_per_task: 5               # 可选，覆盖默认值
  step_limit: 10                   # 可选，预期最大步骤数
```

## 环境变量展开

在任何字符串值中使用 `${ENV_VAR}` 语法：

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}
```

## 默认值级联

套件级别的默认值会应用于未覆盖它们的任务：

```yaml
defaults:
  trials_per_task: 3
  pass_threshold: 0.7
  graders:
    - type: exact_match

tasks:
  - id: task-1
    input:
      prompt: "..."
    # 继承：trials_per_task=3, pass_threshold=0.7, graders=[exact_match]

  - id: task-2
    input:
      prompt: "..."
    trials_per_task: 5              # 覆盖默认值
    graders:                        # 覆盖默认评分器
      - type: contains
```
