import { get, post, del } from './client'
import type { Project } from '@/types'

export interface ProjectInfo {
  name: string
  path: string
  db_path: string
}

export function listProjects(): Promise<Project[]> {
  return get<Project[]>('/projects')
}

export function addProject(name: string, path: string): Promise<Project> {
  return post<Project>('/projects', { name, path })
}

export function deleteProject(name: string): Promise<void> {
  return del<void>(`/projects/${encodeURIComponent(name)}`)
}

export function getProjectInfo(name: string): Promise<ProjectInfo> {
  return get<ProjectInfo>(`/projects/${encodeURIComponent(name)}/info`)
}
