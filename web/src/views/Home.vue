<script setup lang="ts">
import { useRouter } from 'vue-router'
import { NButton, NCard, NSpace, NText } from 'naive-ui'

const router = useRouter()

function goToMemories() {
  router.push({ name: 'memories' })
}

function goCategories() {
  router.push({ name: 'categories' })
}

function goImport() {
  router.push({ name: 'import' })
}
</script>

<template>
  <div class="home-page">
    <div class="welcome-section">
      <h1 class="title">cangjie-mem</h1>
      <p class="subtitle">仓颉语言记忆管理系统</p>
      <p class="description">
        存储和组织你的仓颉语言学习笔记、代码片段和最佳实践
      </p>
    </div>

    <div class="quick-actions">
      <NCard title="快速开始" class="action-card">
        <NSpace vertical size="large">
          <NButton type="primary" size="large" block @click="goToMemories">
            📝 浏览记忆
          </NButton>
          <NButton size="large" block @click="goCategories">
            📁 分类浏览
          </NButton>
          <NButton size="large" block @click="goImport">
            📦 导入/导出
          </NButton>
        </NSpace>
      </NCard>
    </div>

    <div class="info-section">
      <NText depth="3">
        <p>提示：使用左侧导航栏在不同功能之间切换</p>
      </NText>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 60px 24px;
  text-align: center;
}

.welcome-section {
  margin-bottom: 48px;
}

.title {
  font-size: 48px;
  font-weight: 700;
  color: #18a058;
  margin: 0 0 16px 0;
}

.subtitle {
  font-size: 24px;
  color: #333;
  margin: 0 0 16px 0;
}

.description {
  font-size: 16px;
  color: #666;
  margin: 0;
  line-height: 1.6;
}

.quick-actions {
  margin-bottom: 32px;
}

.action-card {
  text-align: left;
}

.info-section {
  padding: 24px;
  background: #f5f7fa;
  border-radius: 8px;
}

.info-section p {
  margin: 0;
}
</style>
