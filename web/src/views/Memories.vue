<script setup lang="ts">
import { useRouter } from 'vue-router'
import { NPageHeader } from 'naive-ui'
import MemoryList from '../components/MemoryList.vue'

const router = useRouter()

function handleBack() {
  router.push({ name: 'home' })
}
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
