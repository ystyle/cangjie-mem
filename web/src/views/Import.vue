<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  NPageHeader, NCard, NText, NSpace, NButton, NUpload, NUploadFileInfo,
  NAlert, NProgress, NTag, NList, NListItem, NSpin, NEmpty, NTabs, NTabPane,
  NCheckbox, NCheckboxGroup, NSelect, NInputGroup, NDivider, NInput
} from 'naive-ui'
import { UploadOutlined, DownloadOutlined, DescriptionOutlined, CheckCircleOutlined, WarningOutlined } from '@vicons/material'
import { NIcon } from 'naive-ui'
import * as api from '../api'
import type { KnowledgePackage, ImportPreview, ImportResult } from '../types'

const router = useRouter()

// 状态
const activeTab = ref('import')
const uploadedFile = ref<File | null>(null)
const importPreview = ref<ImportPreview | null>(null)
const importId = ref('')
const isImporting = ref(false)
const importResult = ref<ImportResult | null>(null)
const uploadError = ref('')

// 导出状态
const exportLevel = ref<string | null>(null)
const exportLibrary = ref<string | null>(null)
const exportProject = ref<string | null>(null)
const isExporting = ref(false)

// 层级选项
const levelOptions = [
  { label: '语言级', value: 'language' },
  { label: '项目级', value: 'project' },
  { label: '库级', value: 'library' },
]

function handleBack() {
  router.push({ name: 'home' })
}

// 文件上传处理
function handleFileChange(options: { fileList: NUploadFileInfo[] }) {
  const file = options.fileList[0]
  if (file && file.file) {
    uploadedFile.value = file.file
    uploadError.value = ''
    importPreview.value = null
    importResult.value = null
    previewImport(file.file)
  }
}

// 预览导入
async function previewImport(file: File) {
  try {
    const text = await file.text()
    const data: KnowledgePackage = JSON.parse(text)

    // 验证格式
    if (!data.version || !data.package || !Array.isArray(data.memories)) {
      throw new Error('文件格式不正确，缺少必需的字段')
    }

    // 调用预览 API
    const preview = await api.previewImport(data)
    importPreview.value = preview
    importId.value = preview.import_id
  } catch (error) {
    uploadError.value = error instanceof Error ? error.message : '文件解析失败'
    console.error('Preview import error:', error)
  }
}

// 确认导入
async function confirmImport() {
  if (!importId.value) return

  isImporting.value = true
  try {
    const result = await api.confirmImport(importId.value)
    importResult.value = result
    // 导入成功后清空预览
    importPreview.value = null
    uploadedFile.value = null
  } catch (error) {
    uploadError.value = error instanceof Error ? error.message : '导入失败'
  } finally {
    isImporting.value = false
  }
}

// 取消导入
function cancelImport() {
  uploadedFile.value = null
  importPreview.value = null
  importResult.value = null
  importId.value = ''
  uploadError.value = ''
}

// 导出记忆
async function handleExport() {
  isExporting.value = true
  try {
    await api.exportMemories({
      level: exportLevel.value || undefined,
      library_name: exportLibrary.value || undefined,
      project_path_pattern: exportProject.value || undefined,
    })
  } catch (error) {
    console.error('Export error:', error)
  } finally {
    isExporting.value = false
  }
}

// 计算导入统计
const importStats = computed(() => {
  if (!importPreview.value) return null
  return {
    total: importPreview.value.total,
    toAdd: importPreview.value.to_add,
    toUpdate: importPreview.value.to_update,
    conflicts: importPreview.value.conflicts.length,
  }
})
</script>

<template>
  <div class="import-page">
    <NPageHeader
      title="导入/导出"
      subtitle="管理知识包的导入和导出"
      @back="handleBack"
      class="page-header"
    />

    <div class="page-content">
      <NTabs v-model:value="activeTab" type="segment" animated>
        <!-- 导入标签页 -->
        <NTabPane name="import" tab="导入知识包">
          <NSpace vertical size="large">
            <!-- 上传区域 -->
            <NCard title="1. 选择文件" size="small">
              <NUpload
                :show-file-list="false"
                accept=".json"
                @change="handleFileChange"
              >
                <NButton secondary>
                  <template #icon>
                    <NIcon :component="UploadOutlined" />
                  </template>
                  选择 JSON 文件
                </NButton>
              </NUpload>

              <NText v-if="uploadedFile" depth="3" style="margin-left: 12px">
                已选择: {{ uploadedFile.name }}
              </NText>

              <NAlert v-if="uploadError" type="error" :title="uploadError" style="margin-top: 12px" />
            </NCard>

            <!-- 预览区域 -->
            <NCard v-if="importPreview" title="2. 预览导入" size="small">
              <NSpace vertical size="medium">
                <!-- 统计信息 -->
                <NAlert type="info">
                  <template #header>
                    导入统计
                  </template>
                  <NSpace>
                    <NTag type="info">总计: {{ importPreview.total }}</NTag>
                    <NTag type="success">新增: {{ importPreview.to_add }}</NTag>
                    <NTag type="warning">更新: {{ importPreview.to_update }}</NTag>
                  </NSpace>
                </NAlert>

                <!-- 冲突列表 -->
                <div v-if="importPreview.conflicts.length > 0">
                  <NText strong style="margin-bottom: 8px; display: block">
                    将被覆盖的记录 ({{ importPreview.conflicts.length }})
                  </NText>
                  <NList bordered size="small" style="max-height: 200px; overflow-y: auto">
                    <NListItem v-for="conflict in importPreview.conflicts" :key="conflict.existing_id">
                      <template #prefix>
                        <NIcon :component="WarningOutlined" color="#f0a020" />
                      </template>
                      <div>
                        <NText>{{ conflict.title }}</NText>
                        <NText depth="3" style="font-size: 12px; margin-left: 8px">
                          {{ conflict.library_name || conflict.level }}
                        </NText>
                      </div>
                    </NListItem>
                  </NList>
                </div>

                <!-- 操作按钮 -->
                <NSpace>
                  <NButton type="primary" :loading="isImporting" @click="confirmImport">
                    <template #icon>
                      <NIcon :component="CheckCircleOutlined" />
                    </template>
                    确认导入
                  </NButton>
                  <NButton @click="cancelImport" :disabled="isImporting">
                    取消
                  </NButton>
                </NSpace>
              </NSpace>
            </NCard>

            <!-- 导入结果 -->
            <NCard v-if="importResult" title="导入完成" size="small">
              <NAlert type="success">
                <template #header>
                  <NIcon :component="CheckCircleOutlined" />
                  导入成功！
                </template>
                <NSpace>
                  <NTag>新增: {{ importResult.added }}</NTag>
                  <NTag>更新: {{ importResult.updated }}</NTag>
                  <NTag type="info">总计: {{ importResult.total }}</NTag>
                </NSpace>
              </NAlert>
            </NCard>

            <!-- 空状态 -->
            <NEmpty v-if="!uploadedFile && !importResult" description="请选择要导入的 JSON 知识包文件">
              <template #icon>
                <NIcon :component="DescriptionOutlined" size="48" />
              </template>
              <template #extra>
                <NText depth="3">
                  文件格式示例：
                </NText>
                <pre style="text-align: left; background: #f5f5f5; padding: 12px; border-radius: 4px; margin-top: 8px; font-size: 12px; overflow-x: auto;">{
  "version": "1.0",
  "package": {
    "name": "知识包名称",
    "description": "知识包描述",
    "version": "1.0.0"
  },
  "memories": [
    {
      "level": "language",
      "title": "记忆标题",
      "content": "记忆内容"
    }
  ]
}</pre>
              </template>
            </NEmpty>
          </NSpace>
        </NTabPane>

        <!-- 导出标签页 -->
        <NTabPane name="export" tab="导出知识包">
          <NSpace vertical size="large">
            <!-- 导出选项 -->
            <NCard title="导出设置" size="small">
              <NSpace vertical size="medium">
                <div>
                  <NText strong style="margin-bottom: 8px; display: block">
                    筛选条件（可选）
                  </NText>
                  <NSpace vertical size="small">
                    <NSelect
                      v-model:value="exportLevel"
                      :options="levelOptions"
                      placeholder="按层级筛选"
                      clearable
                    />
                    <NInput
                      v-model:value="exportLibrary"
                      placeholder="按库名筛选（可选）"
                      clearable
                    />
                    <NInput
                      v-model:value="exportProject"
                      placeholder="按项目路径筛选（可选）"
                      clearable
                    />
                  </NSpace>
                </div>

                <NDivider />

                <NSpace>
                  <NButton
                    type="primary"
                    :loading="isExporting"
                    @click="handleExport"
                  >
                    <template #icon>
                      <NIcon :component="DownloadOutlined" />
                    </template>
                    导出为 JSON
                  </NButton>
                  <NButton @click="() => { exportLevel = null; exportLibrary = null; exportProject = null }">
                    清空筛选
                  </NButton>
                </NSpace>

                <NAlert type="info" style="margin-top: 12px">
                  导出的文件将自动下载，文件名包含时间戳。
                  如果没有设置筛选条件，将导出所有记忆。
                </NAlert>
              </NSpace>
            </NCard>

            <!-- 导出说明 -->
            <NCard title="使用说明" size="small">
              <NSpace vertical size="small">
                <NText>
                  • 导出功能会生成 JSON 格式的知识包文件
                </NText>
                <NText>
                  • 可以使用筛选条件导出特定层级的记忆
                </NText>
                <NText>
                  • 导出的文件可以分享给其他人或备份到本地
                </NText>
                <NText>
                  • 导入时会自动处理冲突（同库同标题的记录会被覆盖）
                </NText>
              </NSpace>
            </NCard>
          </NSpace>
        </NTabPane>
      </NTabs>
    </div>
  </div>
</template>

<style scoped>
.import-page {
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
  max-width: 800px;
  margin: 0 auto;
}

pre {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}
</style>
