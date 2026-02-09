/**
 * API 客户端
 * 封装所有后端 API 调用
 */

import type {
  Memory,
  StoreRequest,
  CategoryInfo,
  KnowledgePackage,
  ImportPreview,
  ImportResult,
} from '../types'

// API 基础 URL
const API_BASE = '/api'

// Basic Auth 凭据（从环境变量或配置读取）
const API_USERNAME = import.meta.env.VITE_API_USERNAME || ''
const API_PASSWORD = import.meta.env.VITE_API_PASSWORD || ''

// API 响应类型
interface APIResponse<T = any> {
  success: boolean
  data?: T
  error?: {
    message: string
    code?: string
  }
}

// API 错误类
export class APIError extends Error {
  code?: string
  status?: number

  constructor(message: string, code?: string, status?: number) {
    super(message)
    this.name = 'APIError'
    this.code = code
    this.status = status
  }
}

// 请求选项
interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: any
  headers?: Record<string, string>
}

// 通用请求函数
async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`
  const { method = 'GET', body, headers = {} } = options

  const config: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
  }

  // 添加 Basic Auth（如果配置了）
  if (API_USERNAME && API_PASSWORD) {
    const credentials = btoa(`${API_USERNAME}:${API_PASSWORD}`)
    config.headers = {
      ...config.headers,
      'Authorization': `Basic ${credentials}`,
    }
  }

  if (body) {
    config.body = JSON.stringify(body)
  }

  try {
    const response = await fetch(url, config)
    const data: APIResponse<T> = await response.json()

    if (!response.ok) {
      throw new APIError(
        data.error?.message || `HTTP ${response.status}`,
        data.error?.code,
        response.status
      )
    }

    if (!data.success) {
      throw new APIError(data.error?.message || '请求失败', data.error?.code)
    }

    return data.data as T
  } catch (error) {
    if (error instanceof APIError) {
      throw error
    }
    throw new APIError(error instanceof Error ? error.message : '网络错误')
  }
}

// ========== 记忆相关 API ==========

// 获取记忆列表
export async function getMemories(params: {
  level?: string
  library_name?: string
  project_path_pattern?: string
  language_tag?: string
  limit?: number
  offset?: number
  order_by?: string
  brief?: boolean
}) {
  const query = new URLSearchParams()
  if (params.level) query.set('level', params.level)
  if (params.library_name) query.set('library_name', params.library_name)
  if (params.project_path_pattern) query.set('project_path_pattern', params.project_path_pattern)
  if (params.language_tag) query.set('language_tag', params.language_tag)
  if (params.limit) query.set('limit', params.limit.toString())
  if (params.offset) query.set('offset', params.offset.toString())
  if (params.order_by) query.set('order_by', params.order_by)
  if (params.brief) query.set('brief', 'true')

  try {
    const result = await request<ListResponse>(`/memories?${query.toString()}`)
    console.log('getMemories response:', result)

    // 将 RecallResult 转换为 Memory
    const memories = (result.results || []).map(recallResultToMemory)

    return {
      total: result.total,
      results: memories,
    }
  } catch (error) {
    console.error('getMemories error:', error)
    throw error
  }
}

interface ListResponse {
  total: number
  results: RecallResult[]
}

// 后端返回的 RecallResult 类型
interface RecallResult {
  id: number
  level: string
  title: string
  content: string
  summary?: string
  library_name?: string
  project_path_pattern?: string
  source: string
  confidence: number
  access_count: number
  matched_text?: string
  created_at?: string
  updated_at?: string
}

// 将 RecallResult 转换为 Memory 的函数
function recallResultToMemory(result: RecallResult): Memory {
  return {
    id: result.id,
    level: result.level as any,
    language_tag: 'cangjie',
    library_name: result.library_name,
    project_path_pattern: result.project_path_pattern,
    title: result.title,
    content: result.content,
    summary: result.summary,
    source: result.source as any,
    access_count: result.access_count,
    confidence: result.confidence,
    created_at: result.created_at || new Date().toISOString(),
    updated_at: result.updated_at || new Date().toISOString(),
    last_accessed_at: undefined,
  }
}

interface StoreResponse {
  success: boolean
  id: number
  message: string
}

interface RecallResponse {
  total: number
  results: Array<{
    id: number
    level: string
    title: string
    content: string
    summary?: string
    library_name?: string
    project_path_pattern?: string
    source: string
    confidence: number
    access_count: number
    created_at?: string
    updated_at?: string
  }>
  search_strategy: string
}

interface CategoriesResponse {
  libraries: CategoryInfo[]
  projects: CategoryInfo[]
}

// 获取单个记忆
export async function getMemory(id: number) {
  return request<Memory>(`/memories/${id}`)
}

// 创建记忆
export async function createMemory(data: StoreRequest) {
  return request<StoreResponse>('/memories', { method: 'POST', body: data })
}

// 更新记忆
export async function updateMemory(id: number, data: StoreRequest) {
  return request<Memory>(`/memories/${id}`, { method: 'PUT', body: data })
}

// 删除记忆
export async function deleteMemory(id: number) {
  return request<void>(`/memories/${id}`, { method: 'DELETE' })
}

// 搜索记忆
export async function searchMemories(data: {
  query: string
  level?: string
  language_tag?: string
  project_context?: string
  max_results?: number
  min_confidence?: number
}) {
  return request<RecallResponse>('/search', { method: 'POST', body: data })
}

// ========== 分类相关 API ==========

// 获取分类统计
export async function getCategories(language_tag?: string) {
  const query = language_tag ? `?language_tag=${language_tag}` : ''
  try {
    const result = await request<CategoriesResponse>(`/categories${query}`)
    console.log('Categories API response:', result)
    return result
  } catch (error) {
    console.error('Categories API error:', error)
    throw error
  }
}

// ========== 导入/导出 API ==========

// 导出记忆
export async function exportMemories(params?: {
  level?: string
  library_name?: string
  project_path_pattern?: string
  language_tag?: string
}) {
  const response = await fetch(`${API_BASE}/export`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params || {}),
  })

  if (!response.ok) {
    throw new APIError('导出失败', undefined, response.status)
  }

  const data = await response.json()
  const blob = new Blob([JSON.stringify(data?.data, null, 2)], { type: 'application/json' })

  // 创建下载链接
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `cangjie-mem-${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  return data
}

// 预览导入
export async function previewImport(pkg: KnowledgePackage) {
  return request<ImportPreview>('/import', { method: 'POST', body: pkg })
}

// 确认导入
export async function confirmImport(importId: string) {
  return request<ImportResult>('/import/confirm', { method: 'POST', body: { import_id: importId } })
}

// ========== 健康检查 ==========

export async function healthCheck() {
  return request<{ status: string }>('/health')
}
