# 高级用法

## CI/CD 集成

在 GitHub Actions 中使用 AgentEval 进行自动化评估：

```yaml
name: Agent Evaluation
on:
  push:
    branches: [main]
jobs:
  eval:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/wallezhang/agent-eval@latest
      - run: agent-eval run -c eval.yaml --fail-under 0.8
```

`--fail-under 0.8` 表示通过率低于 80% 时返回退出码 1，可作为 CI 质量门禁。

## 响应缓存

```yaml
cache:
  enabled: true
  dir: .cache/agent-eval
```

启用后，相同输入的 Agent 响应会被缓存到本地文件系统，避免重复的 API 调用。缓存基于内容寻址，输入相同则直接返回缓存结果。

使用 `--no-cache` 标志可在单次运行中绕过缓存：

```bash
agent-eval run -c eval.yaml --no-cache
```

## 断点续评

评估过程中断后，可通过 `--resume` 从上次中断处继续：

```bash
# 第一次运行被中断
agent-eval run -c eval.yaml
# 使用运行 ID 恢复
agent-eval run -c eval.yaml --resume <run-id>
```

运行 ID 可通过 `agent-eval list` 查看。

## 标签过滤

在任务文件中为任务添加标签：

```yaml
- id: math-addition
  tags: [math, easy]
  input:
    prompt: "What is 2 + 2?"

- id: code-generation
  tags: [coding, hard]
  input:
    prompt: "Write a sorting algorithm"
```

运行时按标签筛选：

```bash
# 只运行 math 标签的任务
agent-eval run -c eval.yaml --tags math
# 排除 hard 标签的任务
agent-eval run -c eval.yaml --exclude-tags hard
```

## 生命周期钩子

```yaml
hooks:
  before_run: "echo '开始评估...'"
  after_run: "curl -X POST https://slack.webhook/... -d '{\"text\": \"评估完成\"}'"
```

钩子在评估前后执行 shell 命令，可用于通知、日志记录等。钩子执行失败不会中断评估（仅记录警告）。

## 自定义 Agent 开发

AgentEval 使用注册模式，添加新的 Agent 类型无需修改现有代码：

1. 在 `pkg/agent/` 下创建新文件（如 `my_agent.go`）
2. 实现 `Agent` 接口（`Execute` + `Close` 方法）
3. 在 `init()` 中注册：`Register("my_agent", factory)`

```go
func init() {
    Register("my_agent", func(config map[string]any) (Agent, error) {
        return &MyAgent{}, nil
    })
}
```

## 自定义评分器开发

与 Agent 相同的模式：

1. 在 `pkg/grader/` 下创建新文件
2. 实现 `Grader` 接口（`Grade` + `Type` 方法）
3. 在 `init()` 中注册

```go
func init() {
    Register("my_grader", func(config map[string]any) (Grader, error) {
        return &MyGrader{}, nil
    })
}
```

## Token 用量与成本估算

AgentEval 自动从 Agent 响应中提取 Token 用量数据，支持两种字段命名：

- OpenAI 格式：`prompt_tokens` / `completion_tokens`
- Anthropic 格式：`input_tokens` / `output_tokens`

Token 数据存储在 `AgentOutput.Metadata["usage"]` 中，报告中会展示总 Token 数和估算成本。
