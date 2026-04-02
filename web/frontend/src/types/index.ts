// Project
export interface Project {
  name: string
  path: string
}

// File tree
export interface FileNode {
  name: string
  type: 'file' | 'dir'
  path: string
  children?: FileNode[]
}

// Config validation
export interface ValidationResult {
  valid: boolean
  errors: string[]
}

// Agent output
export interface AgentOutput {
  text: string
  messages?: Message[]
  metadata?: Record<string, unknown>
}

export interface Message {
  role: string
  content: string
}

// Transcript
export interface TranscriptStep {
  type: string
  role: string
  content: string
  timestamp: string
}

export interface Transcript {
  steps: TranscriptStep[]
}

// Grading
export interface GradeResult {
  grader_type: string
  score: number
  pass: boolean
  weight: number
  reason: string
  error: string
}

// Trial
export type TrialStatus = 'pending' | 'running' | 'passed' | 'failed' | 'error'

export interface Trial {
  id: string
  task_id: string
  index: number
  status: TrialStatus
  agent_output?: AgentOutput
  transcript?: Transcript
  grades: GradeResult[]
  score: number
  pass: boolean
  error: string
  started_at: string
  finished_at: string
  duration_ms: number
  agent_duration_ms: number
  step_count: number
}

// Usage
export interface UsageSummary {
  total_input_tokens: number
  total_output_tokens: number
  total_tokens: number
  estimated_cost_usd: number
}

// Task & Results
export interface TaskResult {
  task: Task
  trials: Trial[]
  pass_count: number
  fail_count: number
  error_count: number
  avg_score: number
  pass_at_k: number
  pass_power_k: number
  latency_p50_ms: number
  latency_p90_ms: number
  latency_p99_ms: number
  avg_step_count: number
  usage?: UsageSummary
}

export interface Task {
  id: string
  name: string
  tags?: string[]
  input: TaskInput
  expected?: ExpectedOutput
  graders: GraderRef[]
  metadata?: Record<string, string>
  trials_per_task: number
}

export interface TaskInput {
  prompt: string
  system?: string
  messages?: Message[]
  params?: Record<string, string>
}

export interface ExpectedOutput {
  text?: string
  json?: Record<string, unknown>
  fields?: Record<string, string>
}

export interface GraderRef {
  type: string
  weight: number
  config?: Record<string, unknown>
}

// Eval run & summary
export interface EvalSummary {
  total_tasks: number
  total_trials: number
  passed_trials: number
  failed_trials: number
  error_trials: number
  overall_pass_rate: number
  avg_score: number
  avg_pass_at_k: number
  avg_pass_power_k: number
  usage?: UsageSummary
}

export interface EvalRun {
  id: string
  suite_name: string
  description: string
  agent_type: string
  agent_config: Record<string, unknown>
  task_results: TaskResult[]
  summary: EvalSummary
  started_at: string
  finished_at: string
  duration_ms: number
}

// Active run (from RunManager)
export interface ActiveRun {
  id: string
  project: string
  started_at: string
}

// SSE events
export interface SSERunStarted {
  run_id: string
  suite: string
  total_tasks: number
}

export interface SSETrialStarted {
  task_id: string
  trial_index: number
}

export interface SSETrialCompleted {
  task_id: string
  trial_index: number
  status: TrialStatus
  score: number
  duration_ms: number
}

export interface SSERunProgress {
  completed: number
  total: number
  pass_count: number
  fail_count: number
  error_count: number
}

export interface SSERunCompleted {
  run_id: string
  summary: EvalSummary
}

export interface SSERunError {
  message: string
}

// Start run request/response
export interface StartRunRequest {
  config_file: string
}

export interface StartRunResponse {
  run_id: string
  status: string
}

// Compare result types
export interface RunMeta {
  id: string
  suite_name: string
  agent_type: string
  started_at: string
}

export interface MetricDiff {
  a: number
  b: number
  diff: number
}

export interface CompareSummary {
  pass_rate: MetricDiff
  avg_score: MetricDiff
  avg_pass_at_k: MetricDiff
  avg_pass_power_k: MetricDiff
}

export interface CompareGradeDetail {
  grader_type: string
  score: number
  pass: boolean
  reason: string
}

export interface CompareTrialDetail {
  status: string
  score: number
  grades: CompareGradeDetail[]
}

export interface TaskComparison {
  task_id: string
  score_a: number
  score_b: number
  diff: number
  status: 'improved' | 'regressed' | 'unchanged'
  trials_a: CompareTrialDetail[]
  trials_b: CompareTrialDetail[]
}

export interface CompareResult {
  run_a: RunMeta
  run_b: RunMeta
  summary: CompareSummary
  tasks: TaskComparison[]
}

// Compare result types
export interface RunMeta {
  id: string
  suite_name: string
  agent_type: string
  started_at: string
}

export interface MetricDiff {
  a: number
  b: number
  diff: number
}

export interface CompareSummary {
  pass_rate: MetricDiff
  avg_score: MetricDiff
  avg_pass_at_k: MetricDiff
  avg_pass_power_k: MetricDiff
}

export interface CompareGradeDetail {
  grader_type: string
  score: number
  pass: boolean
  reason: string
}

export interface CompareTrialDetail {
  status: string
  score: number
  grades: CompareGradeDetail[]
}

export interface TaskComparison {
  task_id: string
  score_a: number
  score_b: number
  diff: number
  status: 'improved' | 'regressed' | 'unchanged'
  trials_a: CompareTrialDetail[]
  trials_b: CompareTrialDetail[]
}

export interface CompareResult {
  run_a: RunMeta
  run_b: RunMeta
  summary: CompareSummary
  tasks: TaskComparison[]
}
