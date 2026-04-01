import { get, post, del } from './client'
import type { EvalRun, ActiveRun, StartRunRequest, StartRunResponse } from '@/types'

export function listRuns(project: string): Promise<EvalRun[]> {
  return get<EvalRun[]>(`/projects/${encodeURIComponent(project)}/runs`)
}

export function listActiveRuns(project: string): Promise<ActiveRun[]> {
  return get<ActiveRun[]>(`/projects/${encodeURIComponent(project)}/runs/active`)
}

export function getRun(project: string, runId: string): Promise<EvalRun> {
  return get<EvalRun>(`/projects/${encodeURIComponent(project)}/runs/${encodeURIComponent(runId)}`)
}

export function startRun(project: string, configFile: string): Promise<StartRunResponse> {
  const body: StartRunRequest = { config_file: configFile }
  return post<StartRunResponse>(`/projects/${encodeURIComponent(project)}/runs`, body)
}

export function deleteRun(project: string, runId: string): Promise<void> {
  return del<void>(`/projects/${encodeURIComponent(project)}/runs/${encodeURIComponent(runId)}`)
}

export function cancelRun(project: string, runId: string): Promise<void> {
  return post<void>(`/projects/${encodeURIComponent(project)}/runs/${encodeURIComponent(runId)}/cancel`)
}

export function getRunSSEUrl(project: string, runId: string): string {
  return `/api/projects/${encodeURIComponent(project)}/runs/${encodeURIComponent(runId)}/sse`
}
