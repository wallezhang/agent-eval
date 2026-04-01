import { get } from './client'

export function listAgentTypes(): Promise<string[]> {
  return get<string[]>('/agents')
}

export function listGraderTypes(): Promise<string[]> {
  return get<string[]>('/graders')
}
