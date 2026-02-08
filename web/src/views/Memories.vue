<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NPageHeader } from 'naive-ui'
import MemoryList from '../components/MemoryList.vue'
import { useAppStore } from '../stores/app'

const router = useRouter()
const appStore = useAppStore()

function handleBack() {
  router.push({ name: 'home' })
}

// 组件挂载时重置 store 筛选条件（让 MemoryList 自己管理筛选）
onMounted(() => {
  // 如果从 Categories 页面跳转过来，清空 store 的筛选条件
  // MemoryList 组件内部有自己的筛选逻辑
  appStore.resetFilters()
})
</script>

<template>
  <div class="memories-page">
    <NPageHeader
      title="记忆列表"
      subtitle="浏览和管理你的仓颉语言记忆"
      @back="handleBack"
      class="page-header"
    />

    <div class="page-content">
      <MemoryList />
    </div>
  </div>
</template>

<style scoped>
.memories-page {
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
