import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  // 侧边栏状态
  const sidebarCollapsed = ref(false)

  // 搜索关键词
  const searchQuery = ref('')

  // 当前选中的分类
  const selectedLevel = ref<string>('')
  const selectedLibrary = ref<string>('')
  const selectedProject = ref<string>('')

  // 方法
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  function setSelectedFilters(level: string, library: string, project: string) {
    selectedLevel.value = level
    selectedLibrary.value = library
    selectedProject.value = project
  }

  function resetFilters() {
    selectedLevel.value = ''
    selectedLibrary.value = ''
    selectedProject.value = ''
    searchQuery.value = ''
  }

  return {
    // 状态
    sidebarCollapsed,
    searchQuery,
    selectedLevel,
    selectedLibrary,
    selectedProject,

    // 方法
    toggleSidebar,
    setSidebarCollapsed,
    setSearchQuery,
    setSelectedFilters,
    resetFilters,
  }
})
