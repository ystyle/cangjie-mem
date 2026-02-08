<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { NPageHeader, NCard, NList, NListItem, NTag, NSpin, NEmpty, NText, NSpace, NButton } from 'naive-ui'
import * as api from '../api'
import { useAppStore } from '../stores/app'
import { useMemoryStore } from '../stores/memory'

const router = useRouter()
const message = useMessage()
const appStore = useAppStore()
const memoryStore = useMemoryStore()

// 数据状态
const loading = ref(true)
const categories = ref<{
  libraries: Array<{ name: string; count: number }>
  projects: Array<{ name: string; count: number }>
}>({ libraries: [], projects: [] })

// 当前选中的分类
const selectedLibrary = ref('')
const selectedProject = ref('')

// 层级标签配置
const levelConfig = {
  language: { label: '语言级', type: 'info' as const },
  project: { label: '项目级', type: 'warning' as const },
  library: { label: '库级', type: 'success' as const },
}

// 加载分类数据
async function loadCategories() {
  loading.value = true
  try {
    const data = await api.getCategories()
    console.log('Loaded categories:', data)
    categories.value = {
      libraries: data.libraries || [],
      projects: data.projects || []
    }
    console.log('Set categories to:', categories.value)
  } catch (error) {
    console.error('Failed to load categories:', error)
    message.error('加载分类失败: ' + (error instanceof Error ? error.message : '未知错误'))
    // 保持初始值，避免 null 错误
    categories.value = { libraries: [], projects: [] }
  } finally {
    loading.value = false
  }
}

// 按库浏览
function handleLibraryClick(libraryName: string) {
  selectedLibrary.value = libraryName
  selectedProject.value = ''
  // 更新 store 的筛选条件并跳转到记忆列表
  appStore.setSelectedFilters('', libraryName, '')
  router.push({ name: 'memories' })
}

// 按项目浏览
function handleProjectClick(projectPath: string) {
  selectedProject.value = projectPath
  selectedLibrary.value = ''
  // 更新 store 的筛选条件并跳转到记忆列表
  appStore.setSelectedFilters('project', '', projectPath)
  router.push({ name: 'memories' })
}

// 按层级浏览
function handleLevelClick(level: string) {
  selectedLibrary.value = ''
  selectedProject.value = ''
  // 更新 store 的筛选条件并跳转到记忆列表
  appStore.setSelectedFilters(level, '', '')
  router.push({ name: 'memories' })
}

// 浏览全部
function handleBrowseAll() {
  selectedLibrary.value = ''
  selectedProject.value = ''
  appStore.resetFilters()
  router.push({ name: 'memories' })
}

function handleBack() {
  router.push({ name: 'home' })
}

// 组件挂载时加载数据
onMounted(() => {
  loadCategories()
})
</script>

<template>
  <div class="categories-page">
    <NPageHeader
      title="分类浏览"
      subtitle="按层级、库或项目浏览记忆"
      @back="handleBack"
      class="page-header"
    />

    <div class="page-content">
      <div v-if="loading" class="loading-container">
        <NSpin size="large" />
        <p>加载中...</p>
      </div>

      <div v-else>
        <!-- 按层级浏览 -->
        <NCard title="按层级浏览" class="card">
          <NSpace :size="12">
            <NTag
              v-for="(_, level) in levelConfig"
              :key="level"
              :type="levelConfig[level].type"
              size="large"
              class="clickable-tag"
              @click="handleLevelClick(level)"
            >
              {{ levelConfig[level].label }}
            </NTag>
            <NTag size="large" class="clickable-tag" @click="handleBrowseAll">
              全部记忆
            </NTag>
          </NSpace>
        </NCard>

        <!-- 按库浏览 -->
        <NCard title="按库浏览" class="card" v-if="categories && categories.libraries && categories.libraries.length > 0">
          <template #header-extra>
            <NText depth="3">共 {{ categories.libraries.length }} 个库</NText>
          </template>

          <NList hoverable clickable>
            <NListItem
              v-for="lib in categories.libraries"
              :key="lib.name"
              @click="handleLibraryClick(lib.name)"
            >
              <template #prefix>
                <span class="icon">📦</span>
              </template>
              <div class="item-content">
                <NText strong>{{ lib.name }}</NText>
                <NTag size="small" round :type="selectedLibrary === lib.name ? 'primary' : 'default'">
                  {{ lib.count }} 条记忆
                </NTag>
              </div>
            </NListItem>
          </NList>

          <NEmpty v-if="categories.libraries.length === 0" description="暂无库级记忆" size="small" />
        </NCard>

        <!-- 按项目浏览 -->
        <NCard title="按项目浏览" class="card" v-if="categories && categories.projects && categories.projects.length > 0">
          <template #header-extra>
            <NText depth="3">共 {{ categories.projects.length }} 个项目</NText>
          </template>

          <NList hoverable clickable>
            <NListItem
              v-for="proj in categories.projects"
              :key="proj.name"
              @click="handleProjectClick(proj.name)"
            >
              <template #prefix>
                <span class="icon">📁</span>
              </template>
              <div class="item-content">
                <NText strong>{{ proj.name }}</NText>
                <NTag size="small" round :type="selectedProject === proj.name ? 'primary' : 'default'">
                  {{ proj.count }} 条记忆
                </NTag>
              </div>
            </NListItem>
          </NList>

          <NEmpty v-if="categories.projects.length === 0" description="暂无项目级记忆" size="small" />
        </NCard>

        <!-- 全空状态 -->
        <NCard v-if="categories && categories.libraries.length === 0 && categories.projects.length === 0">
          <NEmpty description="还没有任何记忆，开始创建你的第一条记忆吧！">
            <template #extra>
              <NButton type="primary" @click="() => router.push({ name: 'memory-new' })">
                创建记忆
              </NButton>
            </template>
          </NEmpty>
        </NCard>
      </div>
    </div>
  </div>
</template>

<style scoped>
.categories-page {
  min-height: 100%;
  background: #f5f7fa;
}

.page-header {
  background: white;
  padding: 16px 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.page-content {
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 16px;
}

.card {
  margin-bottom: 24px;
}

.card:last-child {
  margin-bottom: 0;
}

.clickable-tag {
  cursor: pointer;
  transition: transform 0.2s, opacity 0.2s;
}

.clickable-tag:hover {
  transform: scale(1.05);
  opacity: 0.9;
}

.icon {
  font-size: 20px;
  margin-right: 12px;
}

.item-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex: 1;
}
</style>
