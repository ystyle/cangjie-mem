<script setup lang="ts">
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NButton, NIcon, NDropdown } from 'naive-ui'
import { useAppStore } from '../stores/app'

const router = useRouter()
const appStore = useAppStore()

const menuOptions = [
  {
    label: '记忆列表',
    key: 'memories',
    icon: () => h('span', '📝'),
  },
  {
    label: '分类浏览',
    key: 'categories',
    icon: () => h('span', '📁'),
  },
  {
    label: '导入/导出',
    key: 'import',
    icon: () => h('span', '📦'),
  },
]

const collapsed = computed({
  get: () => appStore.sidebarCollapsed,
  set: (value) => appStore.setSidebarCollapsed(value),
})

function handleMenuSelect(key: string) {
  router.push({ name: key })
}
</script>

<template>
  <NLayout has-sider class="app-layout">
    <NLayoutSider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="200"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="logo">
        <span v-if="!collapsed" class="logo-text">cangjie-mem</span>
        <span v-else class="logo-icon">仓</span>
      </div>

      <NMenu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        @update:value="handleMenuSelect"
      />
    </NLayoutSider>

    <NLayout>
      <NLayoutContent class="content">
        <router-view />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.app-layout {
  height: 100%;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 8px;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: #18a058;
}

.logo-icon {
  font-size: 24px;
  font-weight: 600;
  color: #18a058;
}

.content {
  padding: 24px;
  min-height: 100%;
  background: #f5f7fa;
}
</style>
