import { get, post, put, del } from './client'
import type { ValidationResult } from '@/types'

// encodeConfigPath encodes a config filename for use in URLs.
// Unlike encodeURIComponent, it preserves "/" so sub-paths like "tasks/sample.yaml" work.
function encodeConfigPath(filename: string): string {
  return filename.split('/').map(encodeURIComponent).join('/')
}

export function listConfigs(project: string): Promise<string[]> {
  return get<string[]>(`/projects/${encodeURIComponent(project)}/configs`)
}

export function getConfig(project: string, filename: string): Promise<string> {
  return get<string>(`/projects/${encodeURIComponent(project)}/configs/${encodeConfigPath(filename)}`)
}

export function createConfig(project: string, filename: string, content: string): Promise<void> {
  return post<void>(`/projects/${encodeURIComponent(project)}/configs`, { filename, content })
}

export function updateConfig(project: string, filename: string, content: string): Promise<void> {
  return put<void>(`/projects/${encodeURIComponent(project)}/configs/${encodeConfigPath(filename)}`, { content })
}

export function deleteConfig(project: string, filename: string): Promise<void> {
  return del<void>(`/projects/${encodeURIComponent(project)}/configs/${encodeConfigPath(filename)}`)
}

export function validateConfig(project: string, filename: string): Promise<ValidationResult> {
  return post<ValidationResult>(`/projects/${encodeURIComponent(project)}/configs/validate`, { filename })
}
