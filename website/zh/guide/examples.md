# 示例

## 基础示例：Command Agent

最简单的评估配置，使用 command agent 运行 echo 命令，搭配 exact_match 评分器验证输出。

```yaml
name: "simple-eval"
description: "使用 command agent 的简单评估示例"

agent:
  type: command
  config:
    command: echo
    args: ["Paris"]

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true
        ignore_whitespace: true

execution:
  concurrency: 2
  timeout: 30s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

对应的任务文件 `tasks/capitals.yaml`：

```yaml
- id: france-capital
  name: "法国首都"
  input:
    prompt: "What is the capital of France?"
  expected:
    text: "Paris"
```

## 代码 Agent：OpenAI 评估

使用 OpenAI Agent 评估代码生成能力。`temperature: 0.0` 确保输出的确定性，`rate_limit_rps` 控制 API 调用频率。

```yaml
name: "coding-agent-eval"
description: "代码生成 Agent 评估套件"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1
    temperature: 0.0

defaults:
  trials_per_task: 3
  pass_threshold: 0.5

execution:
  concurrency: 4
  rate_limit_rps: 5
  timeout: 120s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

## 多评分器：组合评分

单个任务使用多个评分器进行综合评估。通过 `weight` 控制各评分器的权重。

示例任务文件：

```yaml
- id: water-formula
  name: "化学式"
  input:
    prompt: "What is the chemical formula for water?"
  expected:
    text: "H2O"
  graders:
    - type: exact_match
      weight: 2.0
    - type: contains
      config:
        keywords: ["H", "O"]
    - type: constraint
      config:
        checks:
          - max_words: 10
```

评分逻辑：

- `exact_match`（权重 2.0）：精确匹配 "H2O"
- `contains`（权重 1.0）：检查输出包含 "H" 和 "O"
- `constraint`（权重 1.0）：确保输出不超过 10 个词
- 最终分数 = 加权平均值
- 通过条件 = 所有评分器都通过

## LLM 评分器：LLM-as-Judge

使用 LLM 作为评判者，基于评分标准（rubric）对 Agent 输出进行打分。适用于难以用规则评估的开放性问题。

```yaml
defaults:
  trials_per_task: 2
  pass_threshold: 0.6
  graders:
    - type: llm
      config:
        provider: openai
        api_key: ${OPENAI_API_KEY}
        model: gpt-4
        rubric: |
          请根据以下标准评估回答：
          1. 正确性
          2. 相关性
          3. 完整性
```

LLM 评分器的响应格式要求包含：`SCORE: <0.0-1.0>`、`PASS: <true/false>`、`REASON: <text>`。
