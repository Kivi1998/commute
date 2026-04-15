<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import CompanyManager from '@/components/CompanyManager.vue'
import CompanyFormModal from '@/components/CompanyFormModal.vue'
import AIRecommendDialog from '@/components/AIRecommendDialog.vue'
import {
  createCompany,
  updateCompany,
  type Company,
  type CompanyCreateInput,
} from '@/api/company'
import { fetchEnums, type EnumItem } from '@/api/meta'

const managerRef = ref<InstanceType<typeof CompanyManager> | null>(null)
const statusOptions = ref<EnumItem[]>([])
const categoryOptions = ref<EnumItem[]>([])

const modalState = reactive<{
  open: boolean
  mode: 'create' | 'edit'
  record: Company | null
}>({
  open: false,
  mode: 'create',
  record: null,
})

const aiDialogOpen = ref(false)

onMounted(async () => {
  const enums = await fetchEnums()
  statusOptions.value = enums.company_status
  categoryOptions.value = enums.company_type
})

function handleAdd() {
  modalState.mode = 'create'
  modalState.record = null
  modalState.open = true
}

function handleEdit(record: Company) {
  modalState.mode = 'edit'
  modalState.record = record
  modalState.open = true
}

async function handleSubmit(payload: CompanyCreateInput) {
  if (modalState.mode === 'create') {
    await createCompany(payload)
    message.success('已添加公司')
  } else if (modalState.record) {
    await updateCompany(modalState.record.id, payload)
    message.success('已更新')
  }
  modalState.open = false
  managerRef.value?.refresh()
}

function openAIDialog() {
  aiDialogOpen.value = true
}

function handleAIImported() {
  managerRef.value?.refresh()
}
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <CompanyManager
      ref="managerRef"
      @add="handleAdd"
      @edit="handleEdit"
      @ai-recommend="openAIDialog"
    />

    <CompanyFormModal
      v-model:open="modalState.open"
      :mode="modalState.mode"
      :record="modalState.record"
      :status-options="statusOptions"
      :category-options="categoryOptions"
      @submit="handleSubmit"
    />

    <AIRecommendDialog v-model:open="aiDialogOpen" @imported="handleAIImported" />
  </div>
</template>
