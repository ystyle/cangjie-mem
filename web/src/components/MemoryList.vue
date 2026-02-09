<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, useDialog, NList, NListItem, NEmpty, NSpin, NButton, NTag, NText, NSpace, NInput, NSelect, NCard, NTooltip, NIcon } from 'naive-ui'
import { SearchOutlined, FilterListOutlined, CloseOutlined, RefreshOutlined, AddCircleOutlined } from '@vicons/material'
import { useMemoryStore } from '../stores/memory'
import { useAppStore } from '../stores/app'
import type { Memory } from '../types'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 配置 dayjs
dayjs.extend(utc)
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const memoryStore = useMemoryStore()
const appStore = useAppStore()

// 本地筛选状态
const searchKeyword = ref('')
const selectedLevel = ref<string | null>(null)
const selectedLibrary = ref<string | null>(null)
const selectedProject = ref<string | null>(null)
const showFilters = ref(false)

// 可用的筛选选项（从 categories 获取）
const availableLibraries = ref<Array<{ label: string; value: string }>>([])
const availableProjects = ref<Array<{ label: string; value: string }>>([])

// 层级配置
const levelConfig = [
  { key: 'language', label: '语言级', icon: '📘', color: '#18a058' as const },
  { key: 'project', label: '项目级', icon: '📁', color: '#f0a020' as const },
  { key: 'library', label: '库级', icon: '📦', color: '#2080f0' as const },
]

// 计算属性
const levelLabels: Record<string, string> = {
  language: '语言',
  project: '项目',
  library: '库',
}

const levelTypes: Record<string, 'info' | 'success' | 'warning'> = {
  language: 'info',
  project: 'warning',
  library: 'success',
}

// 格式化日期
function formatDate(dateStr: string): string {
  const date = dayjs(dateStr)
  const now = dayjs().startOf('day')
  const targetDate = date.startOf('day')

  const diffDays = now.diff(targetDate, 'day')

  let result: string
  if (diffDays === 0) {
    result = '今天'
  } else if (diffDays === 1) {
    result = '昨天'
  } else if (diffDays < 7) {
    result = `${diffDays} 天前`
  } else {
    result = date.format('YYYY-MM-DD')
  }

  return result
}

// 应用筛选条件
async function applyFilters() {
  // 构建完整的查询参数（明确设置所有可能冲突的字段，避免使用 currentParams 中的残留值）
  const params: any = {
    limit: 20,
    offset: 0, // 重置分页
    order_by: 'created_at',
    level: undefined,
    library_name: undefined,
    project_path_pattern: undefined,
  }

  // 只设置当前激活的筛选条件
  if (selectedLevel.value) params.level = selectedLevel.value
  if (selectedLibrary.value) params.library_name = selectedLibrary.value
  if (selectedProject.value) params.project_path_pattern = selectedProject.value

  // 关键词搜索使用 search API
  if (searchKeyword.value.trim()) {
    try {
      const response = await fetch('/api/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: searchKeyword.value.trim(),
          level: selectedLevel.value || undefined,
          max_results: 50,
        }),
      })
      const data = await response.json()
      if (data.success) {
        // 将搜索结果转换为 Memory 格式
        memoryStore.memories = data.data.results.map((r: any) => ({
          id: r.id,
          level: r.level,
          language_tag: 'cangjie',
          title: r.title,
          content: r.content,
          summary: r.summary,
          library_name: r.library_name,
          project_path_pattern: r.project_path_pattern,
          source: r.source,
          access_count: r.access_count,
          confidence: r.confidence,
          created_at: r.created_at || new Date().toISOString(),
          updated_at: r.updated_at || new Date().toISOString(),
        }))
        memoryStore.total = data.data.total
      }
    } catch (error) {
      message.error('搜索失败')
    }
  } else {
    // 无关键词时使用 list API，传递完整的查询参数
    await memoryStore.fetchMemories(params)
  }
}

// 搜索框失去焦点时搜索（如果有内容）
function onSearchBlur() {
  if (searchKeyword.value.trim()) {
    applyFilters()
  }
}

// 清空筛选
function clearFilters() {
  searchKeyword.value = ''
  selectedLevel.value = null
  selectedLibrary.value = null
  selectedProject.value = null
  memoryStore.fetchMemories()
}

// 切换层级筛选
function toggleLevel(level: string) {
  if (selectedLevel.value === level) {
    // 取消选中当前层级
    selectedLevel.value = null
  } else {
    // 选中新层级
    selectedLevel.value = level
    // 清空所有库和项目筛选
    selectedLibrary.value = null
    selectedProject.value = null
  }
  // 手动触发筛选（不依赖 watch）
  applyFilters()
}

// 选择库
function onLibrarySelect(value: string | null) {
  selectedLibrary.value = value
  // 选择库时，自动切换到库级层级
  if (value) {
    selectedLevel.value = 'library'
    selectedProject.value = null
  } else if (selectedLevel.value === 'library') {
    // 如果清空了库选择且当前是库级，也清空层级
    selectedLevel.value = null
  }
  applyFilters()
}

// 选择项目
function onProjectSelect(value: string | null) {
  selectedProject.value = value
  // 选择项目时，自动切换到项目级层级
  if (value) {
    selectedLevel.value = 'project'
    selectedLibrary.value = null
  } else if (selectedLevel.value === 'project') {
    // 如果清空了项目选择且当前是项目级，也清空层级
    selectedLevel.value = null
  }
  applyFilters()
}

// 查看详情
function viewMemory(memory: Memory) {
  router.push({ name: 'memory-edit', params: { id: memory.id } })
}

// 编辑记忆
function editMemory(memory: Memory) {
  router.push({ name: 'memory-edit', params: { id: memory.id } })
}

// 删除记忆
function handleDelete(memory: Memory) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除记忆"${memory.title}"吗？此操作不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await memoryStore.deleteMemory(memory.id)
        message.success('删除成功')
      } catch (error) {
        message.error('删除失败')
      }
    },
  })
}

// 加载更多
function handleLoadMore() {
  memoryStore.loadMore()
}

// 刷新列表
async function handleRefresh() {
  try {
    await memoryStore.fetchMemories()
    message.success('刷新成功')
  } catch (error) {
    message.error('刷新失败')
  }
}

// 新建记忆
function handleNew() {
  router.push({ name: 'memory-new' })
}

// 加载可用的筛选选项
async function loadFilterOptions() {
  try {
    const response = await fetch('/api/categories')
    const data = await response.json()
    if (data.success) {
      availableLibraries.value = (data.data.libraries || []).map((lib: any) => ({
        label: lib.name,
        value: lib.name,
      }))
      availableProjects.value = (data.data.projects || []).map((proj: any) => ({
        label: proj.name,
        value: proj.name,
      }))
    }
  } catch (error) {
    console.error('Failed to load filter options:', error)
  }
}

// 计算激活的筛选数量
const activeFilterCount = computed(() => {
  let count = 0
  if (selectedLevel.value) count++
  if (selectedLibrary.value) count++
  if (selectedProject.value) count++
  if (searchKeyword.value.trim()) count++
  return count
})

// 组件挂载时加载数据
onMounted(async () => {
  console.log('=== MemoryList onMounted ===')
  console.log('appStore:', appStore)
  console.log('appStore.$state:', appStore.$state)
  console.log('typeof appStore.selectedLevel:', typeof appStore.selectedLevel)
  console.log('appStore.selectedLevel value:', appStore.selectedLevel)

  await loadFilterOptions()

  // Pinia store 的状态在 $state 中
  const level = appStore.$state.selectedLevel
  const library = appStore.$state.selectedLibrary
  const project = appStore.$state.selectedProject

  console.log('level from $state:', level)
  console.log('library from $state:', library)
  console.log('project from $state:', project)

  const hasFiltersFromCategories = Boolean(level || library || project)

  console.log('hasFiltersFromCategories:', hasFiltersFromCategories)

  if (hasFiltersFromCategories) {
    // 从 Categories 跳转过来，应用筛选条件
    selectedLevel.value = level || null
    selectedLibrary.value = library || null
    selectedProject.value = project || null

    console.log('Applied filters from appStore:')
    console.log('  selectedLevel:', selectedLevel.value)
    console.log('  selectedLibrary:', selectedLibrary.value)
    console.log('  selectedProject:', selectedProject.value)

    // 清空 appStore 的筛选条件（避免重复应用）
    appStore.resetFilters()

    // 立即应用筛选
    await applyFilters()
  } else {
    console.log('No filters from Categories, resetting all states')

    // 直接访问，重置所有状态
    memoryStore.reset()
    searchKeyword.value = ''
    selectedLevel.value = null
    selectedLibrary.value = null
    selectedProject.value = null

    // 加载记忆列表
    try {
      console.log('Calling fetchMemories() with no filters...')
      await memoryStore.fetchMemories()
    } catch (error) {
      message.error('加载失败')
    }
  }

  console.log('=== MemoryList onMounted end ===')
})
</script>

<template>
  <div class="memory-list">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <!-- 搜索框 -->
      <div class="search-box">
        <NInput
          v-model:value="searchKeyword"
          placeholder="搜索记忆... (按回车搜索)"
          clearable
          size="large"
          @keyup.enter="applyFilters"
          @blur="onSearchBlur"
        >
          <template #prefix>
            <NIcon :component="SearchOutlined" />
          </template>
        </NInput>
      </div>

      <!-- 操作按钮 -->
      <div class="actions">
        <NTooltip>
          <template #trigger>
            <NButton
              circle
              quaternary
              :type="showFilters ? 'primary' : 'default'"
              @click="showFilters = !showFilters"
            >
              <template #icon>
                <NIcon :component="FilterListOutlined" />
              </template>
            </NButton>
          </template>
          筛选{{ activeFilterCount > 0 ? ` (${activeFilterCount})` : '' }}
        </NTooltip>

        <NTooltip>
          <template #trigger>
            <NButton circle quaternary @click="handleRefresh">
              <template #icon>
                <NIcon :component="RefreshOutlined" />
              </template>
            </NButton>
          </template>
          刷新
        </NTooltip>

        <NButton type="primary" circle @click="handleNew">
          <template #icon>
            <NIcon :component="AddCircleOutlined" />
          </template>
        </NButton>
      </div>
    </div>

    <!-- 筛选面板 -->
    <transition name="slide-down">
      <div v-show="showFilters" class="filter-panel">
        <!-- 层级快捷选择 -->
        <div class="filter-section">
          <div class="section-title">快速筛选</div>
          <div class="level-chips">
            <div
              v-for="level in levelConfig"
              :key="level.key"
              :class="['chip', 'level-chip', { active: selectedLevel === level.key }]"
              :style="{ borderColor: selectedLevel === level.key ? level.color : undefined }"
              @click="toggleLevel(level.key)"
            >
              <span class="chip-icon">{{ level.icon }}</span>
              <span class="chip-label">{{ level.label }}</span>
              <NIcon v-if="selectedLevel === level.key" :component="CloseOutlined" size="14" />
            </div>
          </div>
        </div>

        <!-- 高级筛选 -->
        <div v-if="selectedLevel" class="filter-section">
          <div class="section-title">
            {{ selectedLevel === 'library' ? '选择库' : selectedLevel === 'project' ? '选择项目' : '高级选项' }}
          </div>

          <div v-if="selectedLevel === 'library'" class="filter-row">
            <NSelect
              v-model:value="selectedLibrary"
              :options="availableLibraries"
              placeholder="搜索并选择库..."
              clearable
              filterable
              size="small"
              @update:value="onLibrarySelect"
            />
          </div>

          <div v-if="selectedLevel === 'project'" class="filter-row">
            <NSelect
              v-model:value="selectedProject"
              :options="availableProjects"
              placeholder="搜索并选择项目..."
              clearable
              filterable
              size="small"
              @update:value="onProjectSelect"
            />
          </div>
        </div>

        <!-- 清空按钮 -->
        <div v-if="activeFilterCount > 0" class="filter-actions">
          <NButton size="small" quaternary type="error" @click="clearFilters">
            <template #icon>
              <NIcon :component="CloseOutlined" />
            </template>
            清空所有筛选
          </NButton>
        </div>
      </div>
    </transition>

    <!-- 列表内容 -->
    <div v-if="memoryStore.loading && memoryStore.memories.length === 0" class="loading-container">
      <NSpin size="large" />
      <p>加载中...</p>
    </div>

    <NEmpty v-else-if="memoryStore.isEmpty" description="暂无记忆" size="large">
      <template #extra>
        <NButton type="primary" @click="handleNew">
          创建第一条记忆
        </NButton>
      </template>
    </NEmpty>

    <div v-else class="list-container">
      <NList hoverable clickable>
        <NListItem v-for="memory in memoryStore.memories" :key="memory.id">
          <template #prefix>
            <div class="memory-icon">
              <span>{{ levelConfig.find(l => l.key === memory.level)?.icon || '📄' }}</span>
            </div>
          </template>

          <div class="memory-item" @click="viewMemory(memory)">
            <div class="memory-header">
              <NText strong>{{ memory.title }}</NText>
              <NTag :type="levelTypes[memory.level]" size="small" round>
                {{ levelLabels[memory.level] }}
              </NTag>
            </div>

            <NText depth="3" class="memory-summary">
              {{ memory.summary || memory.content.slice(0, 100) + '...' }}
            </NText>

            <div class="memory-footer">
              <NText depth="3" style="font-size: 12px">
                {{ formatDate(memory.created_at) }}
              </NText>
              <NSpace size="small" @click.stop>
                <NButton size="tiny" quaternary @click="editMemory(memory)">
                  编辑
                </NButton>
                <NButton size="tiny" quaternary type="error" @click="handleDelete(memory)">
                  删除
                </NButton>
              </NSpace>
            </div>
          </div>
        </NListItem>
      </NList>

      <div v-if="memoryStore.hasMore" class="load-more">
        <NButton @click="handleLoadMore" :loading="memoryStore.loading" secondary>
          加载更多
        </NButton>
      </div>

      <div v-if="memoryStore.memories.length > 0" class="list-footer">
        共 {{ memoryStore.total }} 条记忆
      </div>
    </div>
  </div>
</template>

<style scoped>
.memory-list {
  max-width: 800px;
  margin: 0 auto;
}

/* 工具栏 */
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
}

.search-box {
  flex: 1;
}

.actions {
  display: flex;
  gap: 8px;
}

/* 筛选面板 */
.filter-panel {
  background: white;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.filter-section {
  margin-bottom: 16px;
}

.filter-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #666;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* 层级芯片 */
.level-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  border: 2px solid transparent;
  background: #f5f5f5;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.chip:hover {
  background: #e8e8e8;
  transform: translateY(-1px);
}

.chip.active {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chip-icon {
  font-size: 16px;
}

.chip-label {
  font-size: 13px;
  font-weight: 500;
}

.filter-row {
  display: flex;
  gap: 12px;
}

.filter-actions {
  display: flex;
  justify-content: center;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 16px;
}

/* 列表容器 */
.list-container {
  background: white;
  border-radius: 12px;
  padding: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.memory-icon {
  font-size: 28px;
  margin-right: 12px;
}

.memory-item {
  flex: 1;
  cursor: pointer;
  padding: 4px 0;
}

.memory-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.memory-summary {
  display: block;
  margin-bottom: 12px;
  line-height: 1.6;
}

.memory-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.load-more {
  text-align: center;
  padding: 16px 0;
}

.list-footer {
  text-align: center;
  padding: 16px 0;
  color: #999;
  font-size: 13px;
}

/* 过渡动画 */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
  max-height: 0;
  margin-bottom: 0;
}

.slide-down-enter-to,
.slide-down-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 300px;
  margin-bottom: 16px;
}
</style>
