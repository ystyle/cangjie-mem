import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as api from '../api'
import type { Memory, StoreRequest, ListRequest } from '../types'

export const useMemoryStore = defineStore('memory', () => {
  // 状态
  const memories = ref<Memory[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 当前查询参数
  const currentParams = ref<Partial<ListRequest>>({
    limit: 20,
    offset: 0,
    order_by: 'created_at',
  })

  // 当前编辑的记忆
  const currentMemory = ref<Memory | null>(null)
  const currentLoading = ref(false)

  // 计算属性（添加防御性检查）
  const hasMore = computed(() => {
    const mems = memories.value || []
    return mems.length < total.value
  })
  const isEmpty = computed(() => {
    const mems = memories.value || []
    return mems.length === 0 && !loading.value
  })

  // 方法
  async function fetchMemories(params?: Partial<ListRequest>, append = false) {
    loading.value = true
    error.value = null

    try {
      const query = { ...currentParams.value, ...params }
      const response = await api.getMemories(query)

      // 防御性检查：确保 results 是数组
      const results = response.results || []

      if (append) {
        memories.value.push(...results)
      } else {
        memories.value = results
      }

      total.value = response.total || 0
      currentParams.value = query
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载失败'
      // 发生错误时，确保保持数组状态
      if (!append) {
        memories.value = []
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function loadMore() {
    if (loading.value || !hasMore.value) return

    const newOffset = (currentParams.value.offset || 0) + (currentParams.value.limit || 20)
    await fetchMemories({ offset: newOffset }, true)
  }

  async function fetchMemory(id: number) {
    currentLoading.value = true
    error.value = null

    try {
      const memory = await api.getMemory(id)
      currentMemory.value = memory
      return memory
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载失败'
      throw err
    } finally {
      currentLoading.value = false
    }
  }

  async function createMemory(data: StoreRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await api.createMemory(data)
      // 刷新列表
      await fetchMemories()
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : '创建失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateMemory(id: number, data: StoreRequest) {
    loading.value = true
    error.value = null

    try {
      const memory = await api.updateMemory(id, data)
      // 更新列表中的项
      const index = memories.value.findIndex(m => m.id === id)
      if (index !== -1) {
        memories.value[index] = memory
      }
      // 更新当前记忆
      if (currentMemory.value?.id === id) {
        currentMemory.value = memory
      }
      return memory
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteMemory(id: number) {
    loading.value = true
    error.value = null

    try {
      await api.deleteMemory(id)
      // 从列表中移除
      memories.value = memories.value.filter(m => m.id !== id)
      total.value--
      // 如果是当前记忆，清空
      if (currentMemory.value?.id === id) {
        currentMemory.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  function setCurrentMemory(memory: Memory | null) {
    currentMemory.value = memory
  }

  function reset() {
    memories.value = []
    total.value = 0
    loading.value = false
    error.value = null
    currentMemory.value = null
    currentParams.value = {
      limit: 20,
      offset: 0,
      order_by: 'created_at',
    }
  }

  return {
    // 状态
    memories,
    total,
    loading,
    error,
    currentMemory,
    currentLoading,
    currentParams,

    // 计算属性
    hasMore,
    isEmpty,

    // 方法
    fetchMemories,
    loadMore,
    fetchMemory,
    createMemory,
    updateMemory,
    deleteMemory,
    setCurrentMemory,
    reset,
  }
})
