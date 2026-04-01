import { get, post } from './client'
import type { FileNode } from '@/types'

export function listFileTree(project: string): Promise<FileNode[]> {
  return get<FileNode[]>(`/projects/${encodeURIComponent(project)}/files`)
}

export function createDir(project: string, path: string): Promise<void> {
  return post<void>(`/projects/${encodeURIComponent(project)}/dirs`, { path })
}
