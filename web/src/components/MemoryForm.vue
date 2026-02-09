<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NForm, NFormItem, NInput, NSelect, NButton, NSpace, NRadioGroup, NRadio } from 'naive-ui'
import { MdEditor, ToolbarNames } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { useMemoryStore } from '../stores/memory'
import type { StoreRequest, KnowledgeLevel } from '../types'

const props = defineProps<{
  memory?: StoreRequest
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'submit', data: StoreRequest): void
  (e: 'cancel'): void
}>()

const router = useRouter()
const message = useMessage()
const memoryStore = useMemoryStore()

// 表单数据
const formData = ref<StoreRequest>({
  level: 'library',
  title: '',
  content: '',
  summary: '',
  language_tag: 'cangjie',
  library_name: '',
  project_path_pattern: '',
  source: 'manual',
})

// 表单引用
const formRef = ref()

// 层级选项
const levelOptions = [
  { label: '公共库级', value: 'library' },
  { label: '项目级', value: 'project' },
  { label: '语言级', value: 'language' },
]

// Markdown 编辑器工具栏配置
const toolbarItems: ToolbarNames[] = [
  'bold',
  'underline',
  'italic',
  'strikeThrough',
  '-',
  'title',
  'sub',
  'sup',
  'quote',
  'unorderedList',
  'orderedList',
  'task',
  '-',
  'codeRow',
  'code',
  'link',
  'image',
  'table',
  '-',
  'revoke',
  'next',
  'save',
  '=',
  'pageFullscreen',
  'fullscreen',
  'preview',
  'htmlPreview',
  'catalog',
]

// 计算属性
const isLibraryLevel = computed(() => formData.value.level === 'library')
const isProjectLevel = computed(() => formData.value.level === 'project')

// 表单验证规则
const rules = {
  title: {
    required: true,
    message: '请输入标题',
    trigger: 'blur',
  },
  content: {
    required: true,
    message: '请输入内容',
    trigger: 'blur',
  },
  library_name: {
    required: isLibraryLevel,
    message: '库级记忆必须指定库名',
    trigger: 'blur',
  },
  project_path_pattern: {
    required: isProjectLevel,
    message: '项目级记忆必须指定项目路径',
    trigger: 'blur',
  },
}

// 监听 props 变化
watch(
  () => props.memory,
  (value) => {
    if (value) {
      formData.value = { ...value }
    }
  },
  { immediate: true }
)

// 层级变化时清空不相关的字段
watch(() => formData.value.level, (newLevel, oldLevel) => {
  if (newLevel === oldLevel) return

  if (newLevel !== 'library') {
    formData.value.library_name = ''
  }
  if (newLevel !== 'project') {
    formData.value.project_path_pattern = ''
  }
})

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate()

    // 验证层级相关字段
    if (isLibraryLevel.value && !formData.value.library_name) {
      message.error('请输入库名')
      return
    }
    if (isProjectLevel.value && !formData.value.project_path_pattern) {
      message.error('请输入项目路径')
      return
    }

    emit('submit', { ...formData.value })
  } catch (error) {
    // 表单验证失败
  }
}

// 取消
function handleCancel() {
  emit('cancel')
}

// 保存并继续
function handleSaveAndContinue() {
  emit('submit', { ...formData.value }, true)
}
</script>

<template>
  <div class="memory-form">
    <NForm
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="top"
      require-mark-placement="right-hanging"
    >
      <!-- 层级选择 -->
      <NFormItem label="层级" path="level">
        <NRadioGroup v-model:value="formData.level">
          <NRadio value="language">语言级</NRadio>
          <NRadio value="project">项目级</NRadio>
          <NRadio value="library">公共库级</NRadio>
        </NRadioGroup>
      </NFormItem>

      <!-- 库名（库级） -->
      <NFormItem v-if="isLibraryLevel" label="库名" path="library_name">
        <NInput
          v-model:value="formData.library_name"
          placeholder="例如：tang, http-client"
          @keydown.enter.prevent="handleSubmit"
        />
      </NFormItem>

      <!-- 项目路径（项目级） -->
      <NFormItem v-if="isProjectLevel" label="项目路径" path="project_path_pattern">
        <NInput
          v-model:value="formData.project_path_pattern"
          placeholder="例如：/path/to/project/*"
          @keydown.enter.prevent="handleSubmit"
        />
      </NFormItem>

      <!-- 标题 -->
      <NFormItem label="标题" path="title">
        <NInput
          v-model:value="formData.title"
          placeholder="简短描述这个知识点"
          @keydown.enter.prevent="handleSubmit"
        />
      </NFormItem>

      <!-- 内容 -->
      <NFormItem label="内容" path="content">
        <MdEditor
          v-model="formData.content"
          placeholder="详细的知识内容，支持 Markdown 格式"
          :toolbars="toolbarItems"
          :style="{ height: '400px' }"
        />
      </NFormItem>

      <!-- 摘要（可选） -->
      <NFormItem label="摘要（可选）">
        <NInput
          v-model:value="formData.summary"
          type="textarea"
          placeholder="简短摘要，用于快速浏览"
          :autosize="{ minRows: 2, maxRows: 4 }"
        />
      </NFormItem>

      <!-- 操作按钮 -->
      <NFormItem>
        <NSpace>
          <NButton type="primary" :loading="loading" @click="handleSubmit">
            保存
          </NButton>
          <NButton @click="handleCancel">
            取消
          </NButton>
        </NSpace>
      </NFormItem>
    </NForm>
  </div>
</template>

<style scoped>
.memory-form {
  background: white;
  padding: 24px;
  border-radius: 8px;
}

.memory-form :deep(.n-form-item-label) {
  font-weight: 500;
}

.memory-form :deep(.n-input),
.memory-form :deep(.n-radio-group) {
  max-width: 100%;
}

/* Markdown 编辑器样式 */
.memory-form :deep(.md-editor) {
  border-radius: 8px;
}

.memory-form :deep(.md-editor-content) {
  min-height: 300px;
}
</style>
