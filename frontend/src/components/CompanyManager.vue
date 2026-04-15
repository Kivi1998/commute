<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { message, Modal, type TableColumnsType } from 'ant-design-vue'
import { PlusOutlined, SearchOutlined, ThunderboltOutlined } from '@ant-design/icons-vue'
import {
  deleteCompany,
  listCompanies,
  updateCompanyStatus,
  type Company,
  type CompanyListQuery,
  type CompanyStatus,
} from '@/api/company'
import { fetchEnums, type EnumItem } from '@/api/meta'
import type { CompanyType } from '@/api/profile'

const emit = defineEmits<{
  (e: 'add'): void
  (e: 'edit', company: Company): void
  (e: 'aiRecommend'): void
}>()

const list = ref<Company[]>([])
const loading = ref(false)
const total = ref(0)

const query = reactive<Required<Pick<CompanyListQuery, 'page' | 'page_size'>> & CompanyListQuery>({
  status: undefined,
  category: undefined,
  keyword: '',
  page: 1,
  page_size: 20,
})

const statusOptions = ref<EnumItem[]>([])
const categoryOptions = ref<EnumItem[]>([])
const sourceOptions = ref<EnumItem[]>([])

const statusColorMap: Record<CompanyStatus, string> = {
  watching: 'default',
  applied: 'blue',
  interviewing: 'orange',
  offered: 'green',
  rejected: 'red',
  archived: 'default',
}

const categoryColorMap: Record<CompanyType, string> = {
  big_tech: 'red',
  mid_tech: 'blue',
  startup: 'purple',
  foreign: 'gold',
  other: 'default',
}

const statusLabel = (v?: string) =>
  statusOptions.value.find((o) => o.value === v)?.label ?? v ?? ''
const categoryLabel = (v?: string) =>
  categoryOptions.value.find((o) => o.value === v)?.label ?? v ?? ''
const sourceLabel = (v?: string) =>
  sourceOptions.value.find((o) => o.value === v)?.label ?? v ?? ''

async function loadEnums() {
  const enums = await fetchEnums()
  statusOptions.value = enums.company_status
  categoryOptions.value = enums.company_type
  sourceOptions.value = enums.company_source
}

async function refresh() {
  loading.value = true
  try {
    const payload: CompanyListQuery = {
      status: query.status,
      category: query.category,
      keyword: query.keyword?.trim() || undefined,
      page: query.page,
      page_size: query.page_size,
    }
    const r = await listCompanies(payload)
    list.value = r.list
    total.value = r.pagination.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  refresh()
}

function handleReset() {
  query.status = undefined
  query.category = undefined
  query.keyword = ''
  query.page = 1
  refresh()
}

async function handleChangeStatus(record: Company, status: CompanyStatus) {
  if (record.status === status) return
  await updateCompanyStatus(record.id, status)
  message.success(`已改为「${statusLabel(status)}」`)
  refresh()
}

function handleDelete(record: Company) {
  Modal.confirm({
    title: `删除「${record.name}」？`,
    content: '删除后不可恢复。',
    okType: 'danger',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      await deleteCompany(record.id)
      message.success('已删除')
      if (list.value.length === 1 && query.page > 1) query.page -= 1
      refresh()
    },
  })
}

function handleTableChange(pagination: { current?: number; pageSize?: number }) {
  query.page = pagination.current ?? 1
  query.page_size = pagination.pageSize ?? 20
  refresh()
}

onMounted(async () => {
  await loadEnums()
  refresh()
})

// 暴露刷新方法给父组件
defineExpose({ refresh })

const columns: TableColumnsType<Company> = [
  { title: '公司', dataIndex: 'name', key: 'name', width: 180 },
  { title: '类型', key: 'category', width: 90 },
  { title: '状态', key: 'status', width: 110 },
  { title: '地址', dataIndex: 'address', key: 'address', ellipsis: true },
  { title: '来源', key: 'source', width: 100 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' },
]

// 当筛选字段变化时重置到第一页
watch(
  () => [query.status, query.category],
  () => handleSearch(),
)
</script>

<template>
  <div>
    <!-- 过滤栏 -->
    <div class="bg-white rounded-md p-3 mb-3 flex flex-wrap gap-3 items-center">
      <a-select
        v-model:value="query.status"
        placeholder="状态"
        allow-clear
        class="!w-36"
        :options="statusOptions.map((o) => ({ value: o.value, label: o.label }))"
      />
      <a-select
        v-model:value="query.category"
        placeholder="类型"
        allow-clear
        class="!w-36"
        :options="categoryOptions.map((o) => ({ value: o.value, label: o.label }))"
      />
      <a-input-search
        v-model:value="query.keyword"
        placeholder="搜索公司名或地址"
        class="!w-72"
        @search="handleSearch"
      >
        <template #enterButton>
          <a-button type="primary">
            <template #icon><SearchOutlined /></template>
          </a-button>
        </template>
      </a-input-search>
      <a-button @click="handleReset">重置</a-button>

      <div class="ml-auto flex gap-2">
        <a-button @click="emit('aiRecommend')">
          <template #icon><ThunderboltOutlined /></template>
          AI 推荐
        </a-button>
        <a-button type="primary" @click="emit('add')">
          <template #icon><PlusOutlined /></template>
          新增公司
        </a-button>
      </div>
    </div>

    <!-- 表格 -->
    <a-table
      :columns="columns"
      :data-source="list"
      :loading="loading"
      row-key="id"
      size="middle"
      :pagination="{
        current: query.page,
        pageSize: query.page_size,
        total,
        showSizeChanger: true,
        pageSizeOptions: ['10', '20', '50'],
        showTotal: (t: number) => `共 ${t} 家公司`,
      }"
      @change="handleTableChange"
    >
      <template #bodyCell="slotProps">
        <template v-if="slotProps.column.key === 'name'">
          <a class="!text-slate-800 font-medium" @click="emit('edit', slotProps.record as Company)">
            {{ slotProps.record.name }}
          </a>
          <div v-if="slotProps.record.industry" class="text-xs text-slate-400 mt-0.5">
            {{ slotProps.record.industry }}
          </div>
        </template>

        <template v-else-if="slotProps.column.key === 'category'">
          <a-tag v-if="slotProps.record.category" :color="categoryColorMap[slotProps.record.category as CompanyType]">
            {{ categoryLabel(slotProps.record.category) }}
          </a-tag>
          <span v-else class="text-slate-300">—</span>
        </template>

        <template v-else-if="slotProps.column.key === 'status'">
          <a-tag :color="statusColorMap[slotProps.record.status as CompanyStatus]">
            {{ statusLabel(slotProps.record.status) }}
          </a-tag>
        </template>

        <template v-else-if="slotProps.column.key === 'source'">
          <span class="text-xs text-slate-500">{{ sourceLabel(slotProps.record.source) }}</span>
        </template>

        <template v-else-if="slotProps.column.key === 'action'">
          <a-space size="small">
            <a-button size="small" type="link" @click="emit('edit', slotProps.record as Company)">编辑</a-button>
            <a-dropdown>
              <a-button size="small" type="link">改状态</a-button>
              <template #overlay>
                <a-menu @click="({ key }: { key: string | number }) => handleChangeStatus(slotProps.record as Company, String(key) as CompanyStatus)">
                  <a-menu-item v-for="opt in statusOptions" :key="opt.value" :disabled="opt.value === slotProps.record.status">
                    {{ opt.label }}
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
            <a-button size="small" type="link" danger @click="handleDelete(slotProps.record as Company)">删除</a-button>
          </a-space>
        </template>
      </template>

      <template #emptyText>
        <a-empty description="还没有公司，点右上角添加" />
      </template>
    </a-table>
  </div>
</template>
