<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, NPageHeader, NSpace } from 'naive-ui'
import { useMemoryStore } from '../stores/memory'
import MemoryForm from '../components/MemoryForm.vue'
import type { StoreRequest } from '../types'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const memoryStore = useMemoryStore()

// 是否为新建模式
const isNew = computed(() => route.name === 'memory-new')

// 记忆 ID
const memoryId = computed(() => {
  const id = route.params.id as string
  return id ? parseInt(id, 10) : null
})

// 当前记忆数据
const currentMemory = computed(() => {
  if (isNew.value) return undefined
  return memoryStore.currentMemory
})

// 加载状态
const loading = computed(() => {
  if (isNew.value) return false
  return memoryStore.currentLoading
})

// 页面标题
const pageTitle = computed(() => isNew.value ? '新建记忆' : '编辑记忆')

// 返回列表
function handleBack() {
  router.push({ name: 'memories' })
}

// 提交表单
async function handleSubmit(data: StoreRequest) {
  try {
    if (isNew.value) {
      await memoryStore.createMemory(data)
      message.success('创建成功')
    } else if (memoryId.value) {
      await memoryStore.updateMemory(memoryId.value, data)
      message.success('更新成功')
    }
    router.push({ name: 'memories' })
  } catch (error) {
    message.error(isNew.value ? '创建失败' : '更新失败')
  }
}

// 取消编辑
function handleCancel() {
  router.push({ name: 'memories' })
}

// 组件挂载时加载数据
onMounted(async () => {
  if (!isNew.value && memoryId.value) {
    try {
      await memoryStore.fetchMemory(memoryId.value)
    } catch (error) {
      message.error('加载失败')
      router.push({ name: 'memories' })
    }
  }
})
</script>

<template>
  <div class="memory-edit-page">
    <NPageHeader
      :title="pageTitle"
      @back="handleBack"
      class="page-header"
    />

    <div class="page-content">
      <MemoryForm
        :memory="currentMemory"
        :loading="loading"
        @submit="handleSubmit"
        @cancel="handleCancel"
      />
    </div>
  </div>
</template>

<style scoped>
.memory-edit-page {
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
}
</style>
