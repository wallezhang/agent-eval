# 快速开始

## 安装

### 一键安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash
```

自动检测操作系统和架构，下载最新二进制文件，验证校验和，安装到 `/usr/local/bin`。可通过 `INSTALL_DIR` 自定义安装目录：

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash
```

### 使用 `go install`

```bash
go install github.com/wallezhang/agent-eval@latest
```

### 下载二进制文件

从 [GitHub Releases](https://github.com/wallezhang/agent-eval/releases) 页面下载预编译的二进制文件。

## 初始化项目

```bash
agent-eval init my-eval
cd my-eval
```

创建的文件结构：

- `eval.yaml` -- 评估配置文件
- `tasks/` -- 任务定义目录
- `tasks/sample.yaml` -- 示例任务文件
- `results/` -- 输出目录

## 编写第一个评估

编辑 `eval.yaml`：

```yaml
name: "my-first-eval"
description: "一个简单的评估"

agent:
  type: command
  config:
    command: echo
    args: ["Hello, World!"]

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true

execution:
  concurrency: 2
  timeout: 30s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

编辑 `tasks/sample.yaml`：

```yaml
- id: greeting
  name: "问候测试"
  input:
    prompt: "Say hello"
  expected:
    text: "Hello, World!"
```

## 运行评估

```bash
agent-eval run -c eval.yaml
```

## 查看结果

结果以表格形式显示在终端中。详细报告保存在 `results/` 目录：

- `results/summary.json` -- 机器可读的结果摘要
- `results/report.html` -- 交互式 HTML 报告

## 查看历史运行

```bash
agent-eval list
```

## 对比两次运行

```bash
agent-eval compare <runA> <runB>
```

## 下一步

- 阅读[核心概念](/zh/guide/concepts)了解评估模型
- 查看[示例](/zh/guide/examples)获取实际配置
- 参阅[配置参考](/zh/reference/configuration)了解所有选项
