# CLI 参考

## agent-eval run

运行评估套件。

```bash
agent-eval run [flags]
```

### 参数

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-c, --config` | string | `eval.yaml` | 评估配置文件路径 |
| `--db` | string | | SQLite 数据库路径 |
| `--verbose` | bool | `false` | 启用详细日志 |
| `--fail-under` | float | `0.0` | 最低通过率（0.0-1.0），低于阈值时退出码为 1 |
| `--tags` | string | | 只运行匹配标签的任务（逗号分隔） |
| `--exclude-tags` | string | | 排除匹配标签的任务（逗号分隔） |
| `--no-cache` | bool | `false` | 本次运行绕过响应缓存 |
| `--resume` | string | | 通过运行 ID 恢复之前的运行 |

### 示例

```bash
# 使用默认配置运行
agent-eval run

# 指定配置文件
agent-eval run -c my-eval.yaml

# CI 模式：通过率低于 80% 时失败
agent-eval run -c eval.yaml --fail-under 0.8

# 只运行 math 标签的任务
agent-eval run --tags math

# 恢复中断的运行
agent-eval run --resume abc123
```

## agent-eval list

列出历史评估运行。

```bash
agent-eval list [flags]
```

### 参数

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--db` | string | `./results/agent-eval.db` | SQLite 数据库路径 |

### 输出

显示包含以下列的表格：
- **ID** -- 运行标识符
- **SUITE** -- 套件名称
- **AGENT** -- Agent 类型
- **TASKS** -- 任务数量
- **PASS RATE** -- 总体通过率
- **DURATION** -- 总运行时长
- **DATE** -- 运行时间戳

### 示例

```bash
agent-eval list
agent-eval list --db ./my-results/agent-eval.db
```

## agent-eval compare

并排对比两次评估运行。

```bash
agent-eval compare <runA> <runB> [flags]
```

### 参数

- `runA` -- 第一次运行的 ID（支持前缀匹配）
- `runB` -- 第二次运行的 ID（支持前缀匹配）

### 标志

| 标志 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--db` | string | `./results/agent-eval.db` | SQLite 数据库路径 |

### 示例

```bash
agent-eval compare abc123 def456
# 前缀匹配
agent-eval compare abc def
```

## agent-eval init

初始化新的评估项目。

```bash
agent-eval init [directory]
```

### 参数

- `directory` -- 目标目录名称（可选，默认为当前目录）

### 创建的文件

```
<directory>/
  eval.yaml          # 评估配置模板
  tasks/
    sample.yaml      # 示例任务文件
  results/           # 输出目录
```

### 示例

```bash
agent-eval init my-eval
cd my-eval
# 编辑 eval.yaml 和 tasks/sample.yaml，然后：
agent-eval run
```
