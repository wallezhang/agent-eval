import { get } from './client'
import type { CompareResult } from '@/types'

export function compareRuns(project: string, runA: string, runB: string): Promise<CompareResult> {
  const params = new URLSearchParams({ runA, runB })
  return get<CompareResult>(`/projects/${encodeURIComponent(project)}/compare?${params.toString()}`)
}
