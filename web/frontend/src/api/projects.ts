import { get, post, del } from './client'
import type { Project } from '@/types'

export function listProjects(): Promise<Project[]> {
  return get<Project[]>('/projects')
}

export function addProject(name: string, path: string): Promise<Project> {
  return post<Project>('/projects', { name, path })
}

export function deleteProject(name: string): Promise<void> {
  return del<void>(`/projects/${encodeURIComponent(name)}`)
}
