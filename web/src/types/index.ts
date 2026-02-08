// 记忆层级
export type KnowledgeLevel = 'language' | 'project' | 'library'

// 记忆来源
export type KnowledgeSource = 'manual' | 'auto_captured'

// 记忆条目
export interface Memory {
  id: number
  level: KnowledgeLevel
  language_tag: string
  library_name?: string
  project_path_pattern?: string
  title: string
  content: string
  summary?: string
  source: KnowledgeSource
  access_count: number
  confidence: number
  created_at: string
  updated_at: string
  last_accessed_at?: string
}

// 列表请求
export interface ListRequest {
  level?: string
  library_name?: string
  project_path_pattern?: string
  language_tag?: string
  limit?: number
  offset?: number
  order_by?: string
  brief?: boolean
}

// 列表响应
export interface ListResponse {
  total: number
  results: RecallResult[]
}

// 回忆结果
export interface RecallResult {
  id: number
  level: KnowledgeLevel
  title: string
  content: string
  summary?: string
  library_name?: string
  project_path_pattern?: string
  source: KnowledgeSource
  confidence: number
  access_count: number
  matched_text?: string
}

// 存储请求
export interface StoreRequest {
  level: KnowledgeLevel
  language_tag?: string
  library_name?: string
  project_path_pattern?: string
  title: string
  content: string
  summary?: string
  source?: KnowledgeSource
}

// 存储响应
export interface StoreResponse {
  success: boolean
  id: number
  message: string
}

// 分类信息
export interface CategoryInfo {
  name: string
  count: number
}

// 分类列表响应
export interface ListCategoriesResponse {
  libraries: CategoryInfo[]
  projects: CategoryInfo[]
}

// 导入请求
export interface ImportRequest {
  memories: Memory[]
  preview?: boolean
}

// 导入预览
export interface ImportPreview {
  total: number
  to_add: number
  to_update: number
  conflicts: ConflictInfo[]
}

// 冲突信息
export interface ConflictInfo {
  title: string
  level: KnowledgeLevel
  library_name?: string
  existing_id: number
  action: 'update' | 'add'
}

// 导入确认请求
export interface ImportConfirmRequest {
  import_id?: string
  strategy: 'skip' | 'overwrite' | 'merge'
}

// 导入响应
export interface ImportResponse {
  total: number
  added: number
  updated: number
  success: boolean
}

// 知识包格式
export interface KnowledgePackage {
  version: string
  package: {
    name: string
    description: string
    author?: string
    tags?: string[]
    version: string
  }
  memories: StoreRequest[]
}

// 回忆响应
export interface RecallResponse {
  total: number
  results: RecallResult[]
  search_strategy: string
}

// 分类响应（简化版）
export type CategoriesResponse = ListCategoriesResponse

// 导入结果
export interface ImportResult {
  added: number
  updated: number
  total: number
}
