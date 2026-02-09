<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, useDialog, NList, NListItem, NEmpty, NSpin, NButton, NTag, NText, NSpace, NInput, NSelect, NCard, NTooltip, NIcon } from 'naive-ui'
import { SearchOutlined, FilterListOutlined, CloseOutlined, RefreshOutlined, AddCircleOutlined, DoneOutlined } from '@vicons/material'
import { useMemoryStore } from '../stores/memory'
import { useAppStore } from '../stores/app'
import MemoryForm from './MemoryForm.vue'
import type { Memory, StoreRequest } from '../types'
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

// 编辑状态
const editingMemory = ref<Memory | null>(null)
const selectedMemoryId = ref<number | null>(null)
const isCreating = ref(false)
const formRef = ref()

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

  if (diffDays === 0) {
    return '今天'
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays} 天前`
  } else {
    return date.format('YYYY-MM-DD')
  }
}

// 检查记忆是否被选中
function isSelected(memory: Memory): boolean {
  return selectedMemoryId.value === memory.id
}

// 选择记忆进行编辑
function selectMemory(memory: Memory) {
  selectedMemoryId.value = memory.id
  editingMemory.value = memory
  isCreating.value = false
}

// 新建记忆
function handleNew() {
  selectedMemoryId.value = null
  editingMemory.value = null
  isCreating.value = true

  // 移动端：滚动到编辑表单
  if (window.innerWidth < 768) {
    nextTick(() => {
      const formSection = document.querySelector('.form-section')
      if (formSection) {
        formSection.scrollIntoView({ behavior: 'smooth', block: 'start' })
      }
    })
  }
}

// 保存记忆（新建或更新）
async function handleSave(data: StoreRequest) {
  try {
    if (isCreating.value) {
      // 新建
      const response = await memoryStore.createMemory(data)
      message.success('创建成功')

      // 保持显示新创建的记忆
      const newMemory: Memory = {
        id: response.id,
        level: data.level,
        language_tag: data.language_tag,
        library_name: data.library_name || '',
        project_path_pattern: data.project_path_pattern || '',
        title: data.title,
        content: data.content,
        summary: data.summary || '',
        source: data.source,
        access_count: 0,
        confidence: 1,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }

      editingMemory.value = newMemory
      selectedMemoryId.value = newMemory.id
      isCreating.value = false

      // 刷新列表
      await applyFilters()
    } else if (editingMemory.value) {
      // 更新
      await memoryStore.updateMemory(editingMemory.value.id, data)
      message.success('保存成功')

      // 更新本地显示
      editingMemory.value = {
        ...editingMemory.value,
        ...data,
      }

      // 刷新列表
      await applyFilters()
    }
  } catch (error) {
    // 错误已在 store 中处理
  }
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

        // 如果删除的是当前编辑的记忆，清空编辑表单
        if (selectedMemoryId.value === memory.id) {
          selectedMemoryId.value = null
          editingMemory.value = null
          isCreating.value = false
        }

        // 刷新列表
        await applyFilters()
      } catch (error) {
        message.error('删除失败')
      }
    },
  })
}

// 应用筛选条件
async function applyFilters() {
  const params: any = {
    limit: 20,
    offset: 0,
    order_by: 'created_at',
    level: undefined,
    library_name: undefined,
    project_path_pattern: undefined,
  }

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
    selectedLevel.value = null
  } else {
    selectedLevel.value = level
    selectedLibrary.value = null
    selectedProject.value = null
  }
  applyFilters()
}

// 选择库
function onLibrarySelect(value: string | null) {
  selectedLibrary.value = value
  if (value) {
    selectedLevel.value = 'library'
    selectedProject.value = null
  } else if (selectedLevel.value === 'library') {
    selectedLevel.value = null
  }
  applyFilters()
}

// 选择项目
function onProjectSelect(value: string | null) {
  selectedProject.value = value
  if (value) {
    selectedLevel.value = 'project'
    selectedLibrary.value = null
  } else if (selectedLevel.value === 'project') {
    selectedLevel.value = null
  }
  applyFilters()
}

// 刷新列表
async function handleRefresh() {
  try {
    await applyFilters()
    message.success('刷新成功')
  } catch (error) {
    message.error('刷新失败')
  }
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
  await loadFilterOptions()

  // 检查是否有从 Categories 页面传来的筛选条件
  const level = appStore.$state.selectedLevel
  const library = appStore.$state.selectedLibrary
  const project = appStore.$state.selectedProject

  const hasFiltersFromCategories = Boolean(level || library || project)

  if (hasFiltersFromCategories) {
    selectedLevel.value = level || null
    selectedLibrary.value = library || null
    selectedProject.value = project || null
    appStore.resetFilters()
    await applyFilters()
  } else {
    memoryStore.reset()
    if (memoryStore.memories.length === 0) {
      try {
        await memoryStore.fetchMemories()
      } catch (error) {
        message.error('加载失败')
      }
    }
  }
})
</script>

<template>
  <div class="memory-list-page">
    <!-- 搜索工具栏 -->
    <div class="toolbar">
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

    <!-- 主内容区：左侧列表，右侧表单 -->
    <div class="content-area">
      <!-- 左侧：记忆列表 -->
      <div class="list-section">
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
            <NListItem
              v-for="memory in memoryStore.memories"
              :key="memory.id"
              :class="['memory-item-wrapper', { selected: isSelected(memory) }]"
              @click="selectMemory(memory)"
            >
              <template #prefix>
                <div class="memory-icon" :class="{ selected: isSelected(memory) }">
                  <span v-if="isSelected(memory)" class="check-icon">
                    <NIcon :component="DoneOutlined" />
                  </span>
                  <span v-else>{{ levelConfig.find(l => l.key === memory.level)?.icon || '📄' }}</span>
                </div>
              </template>

              <div class="memory-item">
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
                    <NButton size="tiny" quaternary type="error" @click="handleDelete(memory)">
                      删除
                    </NButton>
                  </NSpace>
                </div>
              </div>
            </NListItem>
          </NList>

          <div v-if="memoryStore.hasMore" class="load-more">
            <NButton @click="memoryStore.loadMore()" :loading="memoryStore.loading" secondary>
              加载更多
            </NButton>
          </div>

          <div v-if="memoryStore.memories.length > 0" class="list-footer">
            共 {{ memoryStore.total }} 条记忆
          </div>
        </div>
      </div>

      <!-- 右侧：编辑表单 -->
      <div class="form-section">
        <NCard title="编辑记忆" size="small" class="form-card">
          <template #header-extra>
            <NText v-if="editingMemory" depth="3" style="font-size: 12px">
              {{ editingMemory.id ? `ID: ${editingMemory.id}` : '新建' }}
            </NText>
          </template>

          <div class="form-content">
            <MemoryForm
              ref="formRef"
              :memory="editingMemory || undefined"
              :loading="memoryStore.loading"
              @submit="handleSave"
            />
          </div>
        </NCard>
      </div>
    </div>
  </div>
</template>

<style scoped>
.memory-list-page {
  min-height: 100%;
  background: #f5f7fa;
  padding: 24px;
  max-width: 1400px;
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

/* 主内容区 */
.content-area {
  display: grid;
  grid-template-columns: 1fr 1.8fr;
  gap: 20px;
  align-items: start;
}

/* 列表区域 */
.list-section {
  background: white;
  border-radius: 12px;
  padding: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 16px;
}

.memory-item-wrapper {
  transition: background-color 0.2s ease;
  border-radius: 8px;
}

.memory-item-wrapper:hover {
  background-color: #f8f9fa;
}

.memory-item-wrapper.selected {
  background-color: #e6f7ff;
}

.memory-item {
  flex: 1;
  cursor: pointer;
  padding: 4px 0;
}

.memory-icon {
  font-size: 28px;
  margin-right: 12px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.memory-icon.selected {
  background: #2080f0;
  color: white;
}

.check-icon {
  font-size: 18px;
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

/* 表单区域 */
.form-section {
  position: sticky;
  top: 0;
}

.form-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.form-content {
  padding: 16px 0 0 0;
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

/* 响应式布局 */
@media (max-width: 1024px) {
  .content-area {
    grid-template-columns: 1fr 1.5fr;
  }
}

@media (max-width: 768px) {
  .memory-list-page {
    padding: 16px;
  }

  .content-area {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-section {
    position: static;
  }

  .toolbar {
    flex-wrap: wrap;
  }

  .search-box {
    width: 100%;
    order: 1;
  }

  .actions {
    order: 2;
  }
}
</style>
